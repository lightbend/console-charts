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

docs/$(RELEASE).tgz: $(CHART)/* $(CHART)/*/*
	$(call banner)
	helm package $(CHART) -d docs

docs/$(RELEASE_LATEST).tgz: $(CHART)/* $(CHART)/*/*
	$(call banner)
	rm -rf build/$(CHART_LATEST)
	cp -r $(CHART) build/$(CHART_LATEST)
	scripts/munge-to-latest.sh build/$(CHART_LATEST)
	helm package build/$(CHART_LATEST) -d docs

# duplicate method here to generate the index, to avoid pulling in RELEASE as a dependency
latest: init lint docs/es/all-latest.yaml docs/$(RELEASE_LATEST).tgz
	helm repo index docs --url https://lightbend.github.io/helm-charts

clean:
	rm -rf build

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

install-local-latest: docs/$(RELEASE_LATEST).tgz install-helm delete-es
	helm install docs/$(RELEASE_LATEST).tgz --name=es --namespace=lightbend --debug

# always run these steps if in dependencies:
.PHONY: all build install-local install-local-latest install-helm delete-es lint init clean lint-json lint-promql latest
