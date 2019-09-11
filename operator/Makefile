.PHONY: all
all: build check

.PHONY: build
build:
	./build-operator.sh

.PHONY: check
check:
	git diff --summary --exit-code manifests

.PHONY: setup-operator-sdk
setup-operator-sdk:
	./setup-operator-sdk.sh

.PHONY: release
release:
	./release-operator.sh
