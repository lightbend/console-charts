# Using ?= cuz enterprise-suite-latest overrides this
CHART ?= enterprise-suite

RELEASE_NAME = es
NAMESPACE = lightbend

define banner
	$(info === $@)
endef

# Note: We may want to require explicit lists.  As implemented, we may end
# up with random cruft in the packages...
COMPONENTS := $(wildcard */*)
SUBCOMPONENTS := $(wildcard */*/*)

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
build: init package allyaml  ## Build chart package (and all.yaml)

# This target required by top-level helm-charts Makefile.  Will create chart tarball in docs directory
package: $(HELM_CHARTS_DIR)/docs/$(RELEASE).tgz  ## Build chart tarball

script-tests:
	$(CHART_DIR)/tests/lib/bats/bats $(CHART_DIR)/tests
	python $(CHART_DIR)/scripts/lbc_test.py

minikube-tests: build install-local
	@echo Installed Console from helm charts: $(CHART)
	@echo Running tests
	( cd $(CHART_DIR)/tests ; ./smoketest ; )

backend-tests: script-tests minikube-tests

frontend-tests:
	$(CHART_DIR)/scripts/run-e2e-tests.sh

# This target required by top-level helm-charts Makefile.  Tests release on Travis
#
# The conditional bit is so we can control which set of tests are run given we're only
# allowed a single 'test' target.
test:  ## Run tests.  backend-tests by default.  Set env var ES_TEST=frontend to run front end tests.
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

lint-allyaml: allyaml  ## Check that all.yaml is valid (Need Minikube running)
	kubectl --context=minikube create --dry-run --validate -f $(HELM_CHARTS_DIR)/docs/$(RELEASE_NAME)/$(ALL_YAML)

clean::  ## Delete make artifacts
	rm -f $(HELM_CHARTS_DIR)/docs/$(RELEASE_NAME)/$(ALL_YAML)


## These targets used for dev work
#

# Expose services using NodePorts.  Normally disabled but often used
# with minikube.  We enable for debugging purposes.
# (If desired, you can override with something like "make <target> EXPOSE_METHOD=LoadBalancer"
# or "make <target> EXPOSE_METHOD=" to get the default unexposed behavior.)
EXPOSE_SERVICES=--set exposeServices=NodePort
ifneq ($(origin EXPOSE_METHOD), undefined)
  ifeq ($(EXPOSE_METHOD),)
    EXPOSE_SERVICES=
  else
    EXPOSE_SERVICES=--set exposeServices=$(EXPOSE_METHOD)
  endif
endif

allyaml: $(HELM_CHARTS_DIR)/docs/$(RELEASE_NAME)/$(ALL_YAML)  ## Build all.yaml file

$(HELM_CHARTS_DIR)/docs/$(RELEASE_NAME)/$(ALL_YAML): $(HELM_CHARTS_DIR)/docs/$(RELEASE).tgz
	$(call banner)
	mkdir -p $(HELM_CHARTS_DIR)/docs/$(RELEASE_NAME)
	helm --namespace=$(NAMESPACE) template $< $(EXPOSE_SERVICES) > $@

TILLER_NAMESPACE := kube-system

install-helm:  ## Install required helm components into cluster
	-kubectl create namespace $(TILLER_NAMESPACE)
	-kubectl create serviceaccount --namespace $(TILLER_NAMESPACE) tiller
	-kubectl create clusterrolebinding $(TILLER_NAMESPACE):tiller --clusterrole=cluster-admin --serviceaccount=$(TILLER_NAMESPACE):tiller
	-helm init --wait --service-account tiller --upgrade --tiller-namespace=$(TILLER_NAMESPACE)

# helm settings for installing locally
# If set, HELM_VALUES_FILE will be added as '--values' value to install command, overriding default values.
HELM_VALUES_FLAG=$(if $(HELM_VALUES_FILE),--values $(HELM_VALUES_FILE))

# Used for minikube tests.
install-local: $(HELM_CHARTS_DIR)/docs/$(RELEASE).tgz install-helm  ## Install local tarball
	# We set a podUID here for test purposes to ensure everything works as non-root.
	# We usePersistentVolumes to ensure PVCs work as expected.
	-$(CHART_DIR)/scripts/lbc.py install --local-chart=$(HELM_CHARTS_DIR)/docs/$(RELEASE).tgz \
		--wait $(EXPOSE_SERVICES) --set podUID=10001,usePersistentVolumes=true -- $(HELM_VALUES_FLAG)

install-dev: install-helm ## Install content of enterprise-suite/ folder
	-$(CHART_DIR)/scripts/lbc.py install --local-chart=$(CHART_DIR) --wait \
		$(EXPOSE_SERVICES) -- $(HELM_VALUES_FLAG)

help:  ## Print help for targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(lastword $(MAKEFILE_LIST)) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: all build package script-tests minikube-tests test init lint lint-helm lint-json lint-promql lint-allyaml clean \
		allyaml install-helm install-local install-dev help
