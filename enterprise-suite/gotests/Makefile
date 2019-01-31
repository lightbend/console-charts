files := $(shell find . -path ./vendor -prune -o -name '*.go' -print)

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

.PHONY: go-tests
run-tests:
	ginkgo -r -- --namespace=$(TILLER_NAMESPACE) --tiller-namespace=$(TILLER_NAMESPACE)

.PHONY: format
format:
	goimports -w $(files)
