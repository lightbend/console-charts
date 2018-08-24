# Top-level Makefile for working on all helm-charts projects at once.

# Collection of charts to process.
# These are subdirectories of helm-charts.  Add to the list as required.
#
# This setting is the default.  To build a subset (at this level), just define a
# value on the command line.  e.g. "make CHARTS=sample-project"
#
CHARTS = enterprise-suite enterprise-suite-latest reactive-sandbox #sample-project


TOPTARGETS := all build clean test lint-helm minikube-test

$(TOPTARGETS): $(CHARTS)

$(CHARTS):
	$(MAKE) -C $@ $(MAKECMDGOALS)

build-index: docs/index.yaml
	helm repo index docs --url https://lightbend.github.io/helm-charts

docs/index.yaml: $(wildcard docs/*.tgz)

clean:
	rm docs/index.yaml

.PHONY: $(TOPTARGETS) $(CHARTS) build-index clean
