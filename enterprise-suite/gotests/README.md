# Lightbend Console e2e tests

Depends on having ginkgo installed:
`go get -u github.com/onsi/ginkgo/ginkgo`

The gotests should be checked out to `$GOPATH/src/github.com/lightbend/console-charts/enterprise-suite/gotests`.

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

or if console-charts is under GOPATH you can do:
`GO111MODULE=on ginkgo -r`

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

## Dependencies

Dependency tracking is handled by glide, the only tool that seems to properly
support k8s client-go.

See https://github.com/kubernetes/client-go/blob/master/INSTALL.md#glide for how it was set up.

To add a new dependency, add a source file import then run `glide update --strip-vendor`. This
will also update all existing dependencies, unless the value is fixed in `glide.yaml`.
