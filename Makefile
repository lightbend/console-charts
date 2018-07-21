all: init lint build
build: init docs/index.yaml docs/es/all.yaml docs/es/all-latest.yaml

CHART := enterprise-suite
VERSION := $(shell scripts/export-chart-version.sh $(CHART))
RELEASE := $(CHART)-$(VERSION)
CHART_LATEST := $(CHART)-latest
RELEASE_LATEST := $(CHART_LATEST)-$(VERSION)

define banner
	$(info === $@)
endef

docs/es/all.yaml: docs/$(RELEASE).tgz
	$(call banner)
	helm --namespace=lightbend template $< > $@

docs/es/all-latest.yaml: docs/$(RELEASE_LATEST).tgz
	$(call banner)
	helm --namespace=lightbend template $< > $@

docs/index.yaml: docs/$(RELEASE).tgz docs/$(RELEASE_LATEST).tgz
	$(call banner)
	helm repo index docs --url https://lightbend.github.io/helm-charts

docs/$(RELEASE).tgz: $(CHART)/Chart.yaml $(CHART)/templates/*.yaml
	$(call banner)
	helm package $(CHART) -d docs

docs/$(RELEASE_LATEST).tgz: $(CHART)/Chart.yaml $(CHART)/templates/*.yaml
	$(call banner)
	rm -rf build/$(CHART_LATEST)
	cp -r $(CHART) build/$(CHART_LATEST)
	scripts/munge-to-latest.sh build/$(CHART_LATEST)
	helm package build/$(CHART_LATEST) -d docs

clean:
	rm -rf build
	rm -f docs/es/all*
	rm -f docs/$(RELEASE).tgz docs/$(RELEASE_LATEST).tgz

init:
	@scripts/lib.sh
	@helm init -c > /dev/null
	@mkdir -p build

lint: init
	$(call banner)
	helm lint enterprise-suite

lint-json:
	find $(CHART) -name \*.json | xargs -tn 1 jq . >/dev/null

lint-promql:
	./scripts/validate-promql

install-helm:
	-kubectl create serviceaccount --namespace kube-system tiller
	-kubectl create clusterrolebinding tiller-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
	-helm init --service-account tiller

delete-es:
	-helm delete --purge es

install-local: install-helm delete-es
	helm install docs/$(RELEASE).tgz --name=es --namespace=lightbend --debug

# always run these steps if in dependencies:
.PHONY: all build install-local install-helm delete-es lint init clean lint-json lint-promql
