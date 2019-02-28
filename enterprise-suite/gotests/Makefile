files := $(shell find . -path ./vendor -prune -o -name '*.go' -print)

NAMESPACE := lightbend

.PHONY: all
all: check

unformatted = $(shell goimports -l $(files))

.PHONY: check
check: checkformat vet lint

.PHONY: checkformat
checkformat:
	$(if $(unformatted), \
	@echo "needs formatting: $(unformatted)" && \
	echo "run make format" &&\
	exit 1 \
	)

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	@echo skipping `golint -set_exit_status $(go list ./...)`
	@echo disabled due to golint being broken on modules

CLEANUP ?= false

.PHONY: run-tests-minikube
run-tests-minikube:
	ginkgo -r --skip=.*openshift.* --randomizeAllSpecs -- --namespace=$(NAMESPACE) --tiller-namespace=$(NAMESPACE) --cleanup=$(CLEANUP)

.PHONY: run-tests-openshift
run-tests-openshift:
	ginkgo -r --skip=.*minikube.* --randomizeAllSpecs -- --namespace=$(NAMESPACE) --tiller-namespace=$(NAMESPACE) --cleanup=$(CLEANUP)

.PHONY: format
format:
	goimports -w $(files)

.PHONY: setup-tools
setup-tools:
	@echo skipping `go get -u golang.org/x/lint/golint`
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/onsi/ginkgo/ginkgo
