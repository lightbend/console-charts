.PHONY: all
all: build check

.PHONY: build
build:
	scripts/build-operator.sh

.PHONY: check
check:
	git diff --summary --exit-code manifests

.PHONY: setup
setup:
	scripts/setup-operator-sdk.sh

.PHONY: release
release:
	scripts/release-operator.sh
