CHART ?= enterprise-suite
RELEASE_NAME = es

include ../common.mk

ALL_YAML ?= all.yaml

build: $(HELM_CHARTS_DIR)/docs/$(RELEASE_NAME)/$(ALL_YAML)

$(HELM_CHARTS_DIR)/docs/$(RELEASE_NAME)/$(ALL_YAML): $(HELM_CHARTS_DIR)/docs/$(RELEASE).tgz
	$(call banner)
	mkdir -p $(HELM_CHARTS_DIR)/docs/$(RELEASE_NAME)
	helm --namespace=$(NAMESPACE) template $< > $@

test: lint

lint: lint-helm lint-json lint-promql

lint-json:
	find $(CHART_DIR) -name \*.json | xargs -tn 1 jq . >/dev/null

lint-promql:
	$(CHART_DIR)/scripts/validate-promql.sh

minikube-test:
	$(CHART_DIR)/scripts/run-minikube-tests.sh

clean::
	rm -f $(HELM_CHARTS_DIR)/docs/$(RELEASE_NAME)/$(ALL_YAML)

.PHONY: test lint lint-json lint-promql minikube-test
