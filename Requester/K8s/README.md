# How to use this image with K8s
The [pod.yml](pod.yml) file contains an example of how to use the certificate requestor with a pod in K8s. To try it out, run
```
# kubectl apply -f pod.yml
```
The new pod will now be listed as initializing, while it is waiting for the certificate to be approved
```
# kubectl get pod
NAME                             READY     STATUS     RESTARTS   AGE
k8s-certificate-requestor-test   0/1       Init:0/1   0          3s

# kubectl get csr
NAME                             AGE       REQUESTOR                               CONDITION
k8s-certificate-requestor-test   2m        system:serviceaccount:default:default   Pending
```
Approve the certificate by
```
# kubectl certificate approve k8s-certificate-requestor-test
```

The Pod should now be completed, and will have printed the signed certificate details to it's logs :)