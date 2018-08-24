CHART ?= enterprise-suite
RELEASE_NAME = es

include ../common.mk


test: lint

lint: lint-helm lint-json lint-promql

lint-json:
	find $(CHART_DIR) -name \*.json | xargs -tn 1 jq . >/dev/null

lint-promql:
	$(CHART_DIR)/scripts/validate-promql.sh

minikube-test:
	$(CHART_DIR)/scripts/run-minikube-tests.sh

.PHONY: test lint lint-json lint-promql minikube-test
