# Lightbend Console e2e tests

Depends on having ginkgo installed:
`go get github.com/onsi/ginkgo/ginkgo`

Running tests locally with no minikube or helm running:
`ginkgo -r --skip=*openshift* -- --start-minikube`

Running on openshift:
`ginkgo -r --skip=*minikube*`

Running a single test suite:
`ginkgo tests/prometheus`

Test names should include one of prefixes `all:`, `minikube:`, `openshift:` to describe valid platforms.
For example `all:prometheus` should be testing in all environment, while `minikube:ingress` only in minikube environment.
