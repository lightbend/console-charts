.PHONY: all
all: build check

.PHONY: clean
clean:
	rm -rf build

.PHONY: build
build:
	scripts/build-operator.sh

# jsravn: Unfortunately, operator-sdk output is not deterministic, so occasionally
# this will fail erroneously because something in the CSV got swapped around.
# It happens about 10% of the time. Still, this is better than forgetting to
# update the operator.
.PHONY: check
check:
	git --no-pager diff manifests
	git diff --summary --exit-code manifests

.PHONY: setup
setup:
	scripts/setup-operator-sdk.sh
	go get -u github.com/bitnami/kubecfg

.PHONY: release
release:
	scripts/release-operator.sh
