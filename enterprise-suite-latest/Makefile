CHART = enterprise-suite-latest

# !!!!!!!!!!!!!!!!!!!!!!!!!!!!
#
# This is a very atypical chart Makefile.  (aka hack)  It's really just here so we can build
# a 'latest' version of enterprise-suite.
#
# Do not use this as a template for a new chart project.
#
# !!!!!!!!!!!!!!!!!!!!!!!!!!!!

BASE_CHART = enterprise-suite
HELM_CHARTS_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
SCRIPTS_DIR := $(HELM_CHARTS_DIR)/scripts

# Defines before include to tweak behavior as required
VERSION = $(shell $(SCRIPTS_DIR)/export-chart-version.sh $(BASE_CHART))
ALL_YAML = all-latest.yaml

include ../$(BASE_CHART)/Makefile

# post-include overrides
CHART_DIR = ../$(BASE_CHART)
COMPONENTS := $(wildcard ../$(BASE_CHART)/*/.)
SUBCOMPONENTS := $(wildcard ../$(BASE_CHART)/*/*/.)


# Special rule to build 'latest' chart
$(HELM_CHARTS_DIR)/docs/$(filter %-latest,$(CHART))-$(VERSION).tgz: $(COMPONENTS) $(SUBCOMPONENTS)
	$(call banner)
	rm -rf $(HELM_CHARTS_DIR)/build/$(CHART)
	mkdir -p $(HELM_CHARTS_DIR)/build
	cp -r ../$(BASE_CHART) $(HELM_CHARTS_DIR)/build/$(CHART)
	./scripts/munge-to-latest.sh $(HELM_CHARTS_DIR)/build/$(CHART)
	helm package -d $(HELM_CHARTS_DIR)/docs $(HELM_CHARTS_DIR)/build/$(CHART)

clean::
	rm -rf $(HELM_CHARTS_DIR)/build/$(CHART)

.PHONY: clean
