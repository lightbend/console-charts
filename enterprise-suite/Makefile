# Using ?= cuz enterprise-suite-latest overrides this
CHART ?= enterprise-suite

RELEASE_NAME = es
NAMESPACE = lightbend

define banner
	$(info === $@)
endef

# Note: We may want to require explicit lists.  As implemented, we may end
# up with random cruft in the packages...
COMPONENTS := $(wildcard */.)
SUBCOMPONENTS := $(wildcard */*/.)

HELM_CHARTS_DIR := ..
SCRIPTS_DIR := $(HELM_CHARTS_DIR)/scripts

#####
# Note:  These are tweaked by the hacky enterprise-suite-latest Makefile...
VERSION ?= $(shell $(SCRIPTS_DIR)/export-chart-version.sh $(CHART))
RELEASE = $(CHART)-$(VERSION)
CHART_DIR = .
ALL_YAML ?= all.yaml
#####


all: lint build  ## lint then build chart
build: init package $(HELM_CHARTS_DIR)/docs/$(RELEASE_NAME)/$(ALL_YAML)  ## Build chart package (and all.yaml)

# This target required by top-level helm-charts Makefile.  Will create chart tarball in docs directory
package: $(HELM_CHARTS_DIR)/docs/$(RELEASE).tgz

shell-tests:
	$(CHART_DIR)/tests/lib/bats/bats $(CHART_DIR)/tests

minikube-tests:
	$(CHART_DIR)/scripts/run-minikube-tests.sh

backend-tests: shell-tests minikube-tests

frontend-tests:
	$(CHART_DIR)/scripts/run-e2e-tests.sh

# This target required by top-level helm-charts Makefile.  Tests release on Travis
#
# The conditional bit is so we can control which set of tests are run given we're only
# allowed a single 'test' target.
test:
ifneq ($(ES_TEST),frontend)
	$(MAKE) backend-tests
else
	$(MAKE) frontend-tests
endif

# We no longer need a target to build the index.yaml file since that's up to the
# helm-charts Makefile.  Could include this though...
# build: $(HELM_CHARTS_DIR)/docs/index.yaml
#
# $(HELM_CHARTS_DIR)/docs/index.yaml: $(HELM_CHARTS_DIR)/docs/$(RELEASE).tgz
# 	$(call banner)
# 	$(MAKE) -C $(HELM_CHARTS_DIR) index

init:
	@$(SCRIPTS_DIR)/lib.sh
	@helm init -c > /dev/null

# Note: This works for enterprise-suite and allows us to define a latest-specific version in enterprise-suite-latest
$(HELM_CHARTS_DIR)/docs/$(filter-out %-latest,$(CHART))-$(VERSION).tgz: $(COMPONENTS) $(SUBCOMPONENTS)
	$(call banner)
	helm package -d $(HELM_CHARTS_DIR)/docs $(CHART_DIR)

lint: lint-helm lint-json lint-promql  ## Run all lint tests on chart files

lint-helm:  ## Run helm lint on chart files
	helm lint $(CHART_DIR)

lint-json:  ## Test json files are well-formed
	find $(CHART_DIR) -name \*.json | xargs -tn 1 jq . >/dev/null

lint-promql:  ## Test promql bits are well-formed
	$(CHART_DIR)/scripts/validate-promql.sh

clean::  ## Delete make artifacts
	rm -f $(HELM_CHARTS_DIR)/docs/$(RELEASE_NAME)/$(ALL_YAML)


## These targets used for dev work
#

# Install minikube-related services.  These are disabled by default but we enable for debugging purposes.
# (If desired, you can override with "make <target> MINIKUBE=false")
MINIKUBE=true

$(HELM_CHARTS_DIR)/docs/$(RELEASE_NAME)/$(ALL_YAML): $(HELM_CHARTS_DIR)/docs/$(RELEASE).tgz
	$(call banner)
	mkdir -p $(HELM_CHARTS_DIR)/docs/$(RELEASE_NAME)
	helm --namespace=$(NAMESPACE) template $< --set minikube=$(MINIKUBE),internalEnableCredentialsSecret=false > $@

TILLER_NAMESPACE := kube-system

install-helm:  ## Install required helm components into cluster
	-kubectl create namespace $(TILLER_NAMESPACE)
	-kubectl create serviceaccount --namespace $(TILLER_NAMESPACE) tiller
	-kubectl create clusterrolebinding $(TILLER_NAMESPACE):tiller --clusterrole=cluster-admin --serviceaccount=$(TILLER_NAMESPACE):tiller
	-helm init --wait --service-account tiller --tiller-namespace=$(TILLER_NAMESPACE)

delete-local:  ## Delete chart from cluster with helm
	-helm delete --purge $(RELEASE_NAME)

# called from shell script for travis job:
install-local: $(HELM_CHARTS_DIR)/docs/$(RELEASE).tgz install-helm delete-local  ## Install local tarball
	# We set a podUID here for test purposes to ensure everything works as non-root.
	-ES_LOCAL_CHART=$(HELM_CHARTS_DIR)/docs/$(RELEASE).tgz \
		$(CHART_DIR)/install-es.sh --wait --set minikube=$(MINIKUBE),podUID=10001

install-dev: install-helm delete-local  ## Install content of enterprise-suite/ folder
	ES_LOCAL_CHART=$(CHART_DIR) $(CHART_DIR)/install-es.sh --wait --set minikube=$(MINIKUBE)

help:  ## Print help for targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(lastword $(MAKEFILE_LIST)) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: all build package shell-tests minikube-tests test init lint lint-helm lint-json lint-promql clean \
		install-helm delete-local install-local help
