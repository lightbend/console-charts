.PHONY: all
all: build check

.PHONY: build
build:
	./build-operator.sh

.PHONY: check
check:
	scripts/build-operator.sh --validate-diff

.PHONY: setup-operator-sdk
setup-operator-sdk:
	./setup-operator-sdk.sh
