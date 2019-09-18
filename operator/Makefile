.PHONY: all
all: build check

.PHONY: clean
clean:
	rm -rf build

.PHONY: build
build:
	scripts/build-operator.sh

.PHONY: check
check:
	git diff --summary --exit-code manifests

.PHONY: setup
setup:
	scripts/setup-operator-sdk.sh
	go get -u github.com/bitnami/kubecfg

.PHONY: release
release:
	scripts/release-operator.sh
