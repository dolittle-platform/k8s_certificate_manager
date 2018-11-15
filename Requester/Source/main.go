package main

import (
	"crypto/x509/pkix"
	"flag"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
	certificates "k8s.io/api/certificates/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	certutil "k8s.io/client-go/util/cert"
	csrutil "k8s.io/client-go/util/certificate/csr"
)

type stringArray []string

func (v *stringArray) Set(val string) error {
	*v = append(*v, val)
	return nil
}

func (v *stringArray) String() string {
	return strings.Join(*v, " ")
}

// -- Where to store certificates --
var pathPrivateKey string
var pathPublicCert string
var pathPrivateCert string

// -- Certificate Request Details --
var csrCommonName string
var csrSerialNumber string
var csrCountry stringArray
var csrOrganization stringArray
var csrOrganizationalUnit stringArray
var csrLocality stringArray
var csrProvince stringArray
var csrStreetAddress stringArray
var csrPostalCode stringArray

var csrUsage stringArray

const hostnameDefaultValue = "<hostname>"

func defineAndParseFlags() {
	// Flags for where to store the data
	flag.StringVar(&pathPrivateKey, "private-key-path", "/etc/ssl/certs/kubernetes/private.key", "Where to store the private key on disk")
	flag.StringVar(&pathPublicCert, "public-cert-path", "/etc/ssl/certs/kubernetes/public.crt", "Where to store the public signed certificate on disk")
	flag.StringVar(&pathPrivateCert, "private-cert-path", "/etc/ssl/certs/kubernetes/private.crt", "Where to store the file containing both the public certificate and private key on disk")

	// Flags for the CSR subject
	flag.StringVar(&csrCommonName, "common-name", hostnameDefaultValue, "Set the CommonName field of the certificate Subject")
	flag.StringVar(&csrSerialNumber, "serial-number", "", "Set the SerialNumber field of the certificate Subject")
	flag.Var(&csrCountry, "country", "Set the Country field of the certificate Subject (multiple allowed)")
	flag.Var(&csrOrganization, "organization", "Set the Organization field of the certificate Subject (multiple allowed)")
	flag.Var(&csrOrganizationalUnit, "organizational-unit", "Set the OrganizationalUnit field of the certificate Subject (multiple allowed)")
	flag.Var(&csrLocality, "locality", "Set the Locality field of the certificate Subject (multiple allowed)")
	flag.Var(&csrProvince, "province", "Set the Province field of the certificate Subject (multiple allowed)")
	flag.Var(&csrStreetAddress, "street-address", "Set the StreetAddress field of the certificate Subject (multiple allowed)")
	flag.Var(&csrPostalCode, "postal-code", "Set the PostalCode field of the certificate Subject (multiple allowed)")

	flag.Var(&csrUsage, "usage", "Usages to put in the certificate request (multiple allowed)")

	// Override the logtostderr to get the output from the rest of the system
	flag.Set("logtostderr", "true")

	flag.Parse()

	// TODO: Show usage?
}

func setCertificateSubjectDefaults() {
	if csrCommonName == "" || csrCommonName == hostnameDefaultValue {
		// By default, the common name is set to the hostname (which is usually the k8s pod name)
		csrCommonName, _ = os.Hostname()
		// If no organizational unit is set, set it to the k8s namespace
		if csrOrganizationalUnit == nil {
			if nsBytes, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
				csrOrganizationalUnit = []string{string(nsBytes)}
			}
		}
	}
}

func main() {
	defineAndParseFlags()
	setCertificateSubjectDefaults()

	// -- Create K8S client --
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatalln("Error getting kubernetes config:", err)
	}
	client := kubernetes.NewForConfigOrDie(config)

	// -- Generate or reuse the private key --
	privateKeyData, generated, err := certutil.LoadOrGenerateKeyFile(pathPrivateKey)
	if err != nil {
		glog.Fatalln("Error generating private key:", err)
	}
	if generated {
		glog.Infoln("Generated new private key, stored in", pathPrivateKey)
	} else {
		glog.Infoln("Reusing previous private key from", pathPrivateKey)
	}

	privateKey, err := certutil.ParsePrivateKeyPEM(privateKeyData)
	if err != nil {
		glog.Fatalln("Error parsing private key:", err)
	}

	// -- Make the Certificate Signing Request --
	subject := pkix.Name{
		CommonName:         csrCommonName,
		SerialNumber:       csrSerialNumber,
		Country:            csrCountry,
		Organization:       csrOrganization,
		OrganizationalUnit: csrOrganizationalUnit,
		Locality:           csrLocality,
		Province:           csrProvince,
		StreetAddress:      csrStreetAddress,
		PostalCode:         csrPostalCode,
	}
	hostNames := []string{} // TODO: Add something here for services
	csrData, err := certutil.MakeCSR(privateKey, &subject, hostNames, nil)
	if err != nil {
		glog.Fatalln("Error making CSR:", err)
	}

	csrClient := client.CertificatesV1beta1().CertificateSigningRequests()

	usages := []certificates.KeyUsage{}
	for _, usage := range csrUsage {
		usages = append(usages, certificates.KeyUsage(usage))
	}

	csrName := csrCommonName // TODO: This should perhaps be generated by k8s?
	request, err := csrutil.RequestCertificate(csrClient, csrData, csrName, usages, privateKey)
	if err != nil {
		glog.Fatalln("Error creating CSR", csrName, ":", err)
	} else {
		glog.Infoln("Created CSR", csrName, ", waiting for approval...")
	}

	certData, err := csrutil.WaitForCertificate(csrClient, request, time.Hour*1)
	if err != nil {
		glog.Fatalln("Error waiting for the CSR", csrName, "to be signed:", err)
	} else {
		glog.Infoln("CSR", csrName, "signed!")
	}

	err = certutil.WriteCert(pathPublicCert, certData)
	if err != nil {
		glog.Fatalln("Error writing public signed certificate to", pathPublicCert, ", error:", err)
	} else {
		glog.Infoln("Public signed certificate written to", pathPublicCert)
	}

	certKeyData := append(certData, privateKeyData...)
	err = certutil.WriteKey(pathPrivateCert, certKeyData)
	if err != nil {
		glog.Fatalln("Error writing public certificate and private key to", pathPrivateCert, ", error:", err)
	} else {
		glog.Infoln("Public certificate and private key written to", pathPrivateCert)
	}
}
