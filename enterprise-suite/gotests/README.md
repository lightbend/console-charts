# Lightbend Console e2e tests

Depends on having ginkgo installed:
`go get github.com/onsi/ginkgo/ginkgo`

Test suites assume that a Kubernetes cluster is running with Helm Tiller installed, but without Lightbend Console. Appropriate sequence of commands to accomplish that locally with minikube is this:
```
minikube start --cpus=3 --memory=6000
kubectl create serviceaccount --namespace kube-system tiller
kubectl create clusterrolebinding kube-system:tiller --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
helm init --wait --service-account tiller --tiller-namespace=kube-system
```
Alternatively, you can use `go-tests` and `go-tests-start-minikube` targets in the `enterprise-suite/Makefile`.

Running all tests locally:
`ginkgo -r`

If Tiller is installed in other namespace than `kube-system`, you can specify that with a flag:
`ginkgo -r -- --tiller-namespace=lightbend-test`
Note: this is not a ginkgo flag, but a custom flag in the test suites so it comes after `--`.

Running only minikube or openshift:
`ginkgo -r --skip=.*minikube.*`
`ginkgo -r --skip=.*openshift.*`

Running a single test suite:
`ginkgo tests/prometheus`

Test names should include one of prefixes `all:`, `minikube:`, `openshift:` to describe valid platforms.
For example `all:prometheus` should be testing in all platforms, while `minikube:ingress` only in minikube.

Tracking of dependencies is done using Go 1.11 module system. To add a dependency, simply write an import
statement in the code, nothing else needs to be done. Updating is done using `go get`, more about it in the 
[golang wiki](https://github.com/golang/go/wiki/Modules#how-to-upgrade-and-downgrade-dependencies).
