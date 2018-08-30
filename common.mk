# This file should be included from chart-specific Makefiles
# Those Makefiles should define the following values BEFORE including this file:
#
# CHART          # Chart name (must match directory name)
# RELEASE_NAME   # Name of helm release
# NAMESPACE      # Optional.  Namespace used when helming.  Default is 'lightbend'
#

#!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
#
# NOTE:
# Please be careful modifying this file.  Changes here will affect all of the 
# helm-charts projects
#
#!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

ifndef CHART
    $(error CHART is not set)
endif

ifndef RELEASE_NAME
    $(error RELEASE_NAME is not set)
endif

# Default value for namespace used with helm.  Override if required.
NAMESPACE ?= lightbend


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
# Note:  There's no need to override any of these variables.
#   (...except in the hacky enterprise-suite-latest Makefile...)
VERSION ?= $(shell $(SCRIPTS_DIR)/export-chart-version.sh $(CHART))
RELEASE = $(CHART)-$(VERSION)
CHART_DIR = .
#####


all: test build  ## Test then build chart
build: init $(HELM_CHARTS_DIR)/docs/index.yaml  ## Build chart


$(HELM_CHARTS_DIR)/docs/index.yaml: $(HELM_CHARTS_DIR)/docs/$(RELEASE).tgz
	$(call banner)
	helm repo index $(HELM_CHARTS_DIR)/docs --url https://lightbend.github.io/helm-charts

# Note: This will skip a chart whose name ends in "-latest".  Don't do that...
$(HELM_CHARTS_DIR)/docs/$(filter-out %-latest,$(CHART))-$(VERSION).tgz: $(COMPONENTS) $(SUBCOMPONENTS)
	$(call banner)
	helm package -d $(HELM_CHARTS_DIR)/docs $(CHART_DIR)

test: lint-helm  ## Run tests on chart components

lint-helm:  ## Run helm lint on chart files
	helm lint $(CHART_DIR)

init:
	@$(SCRIPTS_DIR)/lib.sh
	@helm init -c > /dev/null

install-helm:  ## Install required helm components into cluster
	-kubectl create serviceaccount --namespace kube-system tiller
	-kubectl create clusterrolebinding tiller-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
	-helm init --wait --service-account tiller

delete-local:  ## Delete chart from cluster with helm
	-helm delete --purge $(RELEASE_NAME)

install-local: $(HELM_CHARTS_DIR)/docs/$(RELEASE).tgz install-helm delete-local  ## Install local chart
	helm install $(HELM_CHARTS_DIR)/docs/$(RELEASE).tgz --name=$(RELEASE_NAME) --namespace=$(NAMESPACE) --debug --wait

clean::  ## Clean up

minikube-test:

help:  ## Print help for targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(lastword $(MAKEFILE_LIST)) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: all build test init lint-helm install-helm delete-local install-local clean minikube-test help
