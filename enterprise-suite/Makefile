# Using ?= cuz enterprise-suite-latest overrides this
CHART ?= enterprise-suite

# Caller can override these with env vars if desired.  e.g. "export NAMESPACE=custom ; make "
RELEASE_NAME ?= es
NAMESPACE ?= lightbend
TILLER_NAMESPACE ?= $(NAMESPACE)

# If set, HELM_VALUES_FILE will be added as '--values' value to install command, overriding default values.
HELM_VALUES_FLAG=$(if $(HELM_VALUES_FILE),--values $(HELM_VALUES_FILE))

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
CONSOLE_API_VERSION := $(shell yq r $(CHART_DIR)/values.yaml esMonitorVersion)
ALL_YAML ?= all.yaml
#####

.PHONY: all build
all: lint build  ## lint then build chart
build: package allyaml  ## Build chart package (and all.yaml)

# This target required by top-level helm-charts Makefile.  Will create chart tarball in build directory.
.PHONY: package
package: $(HELM_CHARTS_DIR)/build/$(RELEASE).tgz  ## Build chart tarball

.PHONY: script-tests
script-tests: ## Run installer tests
	python $(CHART_DIR)/scripts/lbc_test.py

.PHONY: frontend-tests
frontend-tests: ## Run all the frontend tests against minikube
	$(CHART_DIR)/scripts/run-e2e-tests.sh

.PHONY: frontend-tests1
frontend-tests1: ## Run subset 1 of frontend tests against minikube
	$(CHART_DIR)/scripts/run-e2e-tests.sh 1

.PHONY: frontend-tests2
frontend-tests2: ## Run subset 2 of frontend tests against minikube
	$(CHART_DIR)/scripts/run-e2e-tests.sh 2

# This target required by top-level helm-charts Makefile. It's not used by travis.
# It only performs lightweight checks.
.PHONY: test
test: lint script-tests ## Run unit tests

# Note: This works for enterprise-suite and allows us to define a latest-specific version in enterprise-suite-latest
$(HELM_CHARTS_DIR)/build/$(filter-out %-latest,$(CHART))-$(VERSION).tgz: $(COMPONENTS) $(SUBCOMPONENTS)
	$(call banner)
	mkdir -p $(HELM_CHARTS_DIR)/build
	helm package -d $(HELM_CHARTS_DIR)/build $(CHART_DIR)

.PHONY: lint
lint: lint-helm lint-json lint-promql  ## Run all lint tests on chart files

.PHONY: lint-helm
lint-helm:  ## Run helm lint on chart files
	helm lint $(CHART_DIR)

.PHONY: lint-json
lint-json:  ## Test json files are well-formed
	find $(CHART_DIR)/console-api $(CHART_DIR)/es-grafana -name '*.json' | xargs -tn 1 jq . >/dev/null

.PHONY: lint-promql
lint-promql:  ## Test promql bits are well-formed
	$(CHART_DIR)/scripts/validate-promql.sh

.PHONY: lint-allyaml
lint-allyaml: allyaml  ## Check that all.yaml is valid (Need Minikube running)
	kubectl --context=minikube create --dry-run --validate -f $(HELM_CHARTS_DIR)/build/$(RELEASE_NAME)/$(ALL_YAML)

.PHONY: check-default-monitors-sync
check-default-monitors-sync:
	$(CHART_DIR)/scripts/pull-console-api.sh $(CONSOLE_API_VERSION) monitors/default-monitors.json | diff -u $(CHART_DIR)/console-api/default-monitors.json -
	$(CHART_DIR)/scripts/pull-console-api.sh $(CONSOLE_API_VERSION) monitors/static-rules.yml | diff -u $(CHART_DIR)/console-api/static-rules.yml -

.PHONY: sync-default-monitors
sync-default-monitors:
	$(CHART_DIR)/scripts/pull-console-api.sh $(CONSOLE_API_VERSION) monitors/default-monitors.json > $(CHART_DIR)/console-api/default-monitors.json
	$(CHART_DIR)/scripts/pull-console-api.sh $(CONSOLE_API_VERSION) monitors/static-rules.yml > $(CHART_DIR)/console-api/static-rules.yml

.PHONY: clean
clean::  ## Delete make artifacts
	rm -f $(HELM_CHARTS_DIR)/build/$(RELEASE_NAME)/$(ALL_YAML)

.PHONY: allyaml
allyaml: $(HELM_CHARTS_DIR)/build/$(RELEASE_NAME)/$(ALL_YAML)  ## Build all.yaml file

$(HELM_CHARTS_DIR)/build/$(RELEASE_NAME)/$(ALL_YAML): $(HELM_CHARTS_DIR)/build/$(RELEASE).tgz
	$(call banner)
	mkdir -p $(HELM_CHARTS_DIR)/build/$(RELEASE_NAME)
	helm --namespace=$(NAMESPACE) template $< --set exposeServices=NodePort > $@

.PHONY: install-helm
install-helm:  ## Install helm (tiller)
	-kubectl create namespace $(TILLER_NAMESPACE)
	-kubectl create serviceaccount --namespace $(TILLER_NAMESPACE) tiller
	-kubectl create clusterrolebinding $(TILLER_NAMESPACE):tiller --clusterrole=cluster-admin \
		--serviceaccount=$(TILLER_NAMESPACE):tiller
	helm init --wait --service-account tiller --upgrade --tiller-namespace=$(TILLER_NAMESPACE)

.PHONY: install-dev
install-dev:  ## Install local chart directory
	TILLER_NAMESPACE=$(TILLER_NAMESPACE) $(CHART_DIR)/scripts/lbc.py install --local-chart=$(CHART_DIR) --namespace $(NAMESPACE) -- \
	    $(HELM_VALUES_FLAG) \
	    --set exposeServices=NodePort \
	    --wait

.PHONE: preflight-check
preflight-check:
	$(MAKE) -C gotests setup-tools
	$(MAKE) -C gotests check

.PHONY: gotests-minikube
gotests-minikube: preflight-check
	$(MAKE) -C gotests run-tests-minikube NAMESPACE=$(NAMESPACE)

.PHONY: gotests-openshift
gotests-openshift: preflight-check
	# Purge any existing installation first.
	-$(MAKE) -C gotests purge-console-openshift NAMESPACE=$(NAMESPACE)
	$(MAKE) -C gotests run-tests-openshift NAMESPACE=$(NAMESPACE)
	# Clean up after the tests.
	-$(MAKE) -C gotests purge-console-openshift NAMESPACE=$(NAMESPACE)

.PHONY: build-operator
build-operator:
	scripts/build-operator.sh

.PHONY: verify-operator
verify-operator:
	scripts/build-operator.sh --verify-diff

.PHONY: help
help:  ## Print help for targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(lastword $(MAKEFILE_LIST)) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
