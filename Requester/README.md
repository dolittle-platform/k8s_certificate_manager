# K8s certificate requester
The certificate requester is a simple Docker image that upon start:

1. Generates a private key
1. Submits a Certificate Signing Request to the K8s apiserver
1. Waits for the CSR to be approved and issued by the
1. Stores the private key and signed certificate on disk

The typical use for this image, is as an `initContainer` for a K8s Pod, to acquire a certificate that is signed by the K8s cluster Certificate Authority, to be used for authentication among services within the cluster. Such as for a MongoDB replica-set.

The Subject of the requested certificate is configurable through arguments to the container:
- `--common-name` - single value, defaults to the hostname (usuaylly name of the Pod).
- `--serial-number` - single value.
- `--country` - multiple values allowed (repeat the argument).
- `--organization` - multiple values.
- `--organizational-unit` - multiple values, defaults to the K8s namespace.
- `--locality` - multiple values.
- `--province` - multiple values.
- `--street-address` - multiple values.
- `--postal-code` - multiple values.

The certificate Usage can also be configured using the `--usage` argument (multiple allowed). The appropriate values for the argument is defined in the [K8s Go client api](https://godoc.org/k8s.io/api/certificates/v1beta1#KeyUsage).

Additionaly, three paths are used to store files on disk:
- `--private-key-path` - the path to the file where a private key will be read from, or a new one written to if it doesn't exist.
- `--public-cert-path` - the signed certificate will be written to this file.
- `--private-cert-path` - both the signed certificate, and the private key will be written to this file.