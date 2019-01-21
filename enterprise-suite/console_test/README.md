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

Running all tests locally:
`ginkgo -r`

Running only minikube or openshift:
`ginkgo -r --skip=*minikube*`
`ginkgo -r --skip=*openshift*`

Running a single test suite:
`ginkgo tests/prometheus`

Test names should include one of prefixes `all:`, `minikube:`, `openshift:` to describe valid platforms.
For example `all:prometheus` should be testing in all environment, while `minikube:ingress` only in minikube environment.
