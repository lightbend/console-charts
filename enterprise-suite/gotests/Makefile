files := $(shell find . -path ./vendor -prune -o -name '*.go' -print)

NAMESPACE := lightbend
TILLER_NAMESPACE := $(NAMESPACE)

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
	golint -set_exit_status $(go list ./...)

.PHONY: purge-console-openshift
purge-console-openshift: purge-console
	-oc delete route console-server

.PHONY: purge-console
purge-console:
	-TILLER_NAMESPACE=$(TILLER_NAMESPACE) ../scripts/lbc.py uninstall --delete-pvcs

ginkgo_args := -r -compilers=2 --progress --fastFail --randomizeAllSpecs --slowSpecThreshold=30 -- --namespace=$(NAMESPACE) --tiller-namespace=$(TILLER_NAMESPACE)

.PHONY: run-tests-minikube
run-tests-minikube:
	ginkgo --skip=.*openshift.* $(ginkgo_args)

.PHONY: run-tests-openshift
run-tests-openshift:
	ginkgo --skip=.*minikube.* $(ginkgo_args)

.PHONY: format
format:
	goimports -w $(files)

.PHONY: setup-tools
setup-tools:
	go get -u golang.org/x/lint/golint
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/onsi/ginkgo/ginkgo
