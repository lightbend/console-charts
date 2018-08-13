all: init lint build
build: init docs/index.yaml docs/es/all.yaml docs/es/all-latest.yaml

CHART := enterprise-suite
export VERSION := $(shell scripts/export-chart-version.sh $(CHART))
export RELEASE := $(CHART)-$(VERSION)
CHART_LATEST := $(CHART)-latest
export RELEASE_LATEST := $(CHART_LATEST)-$(VERSION)

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

release:
	$(call banner)
	echo "stub: release"

clean:
	rm -rf build

test:
	$(MAKE) -C $(CHART) $@

minikube-test:
	$(MAKE) -C $(CHART) $@

init:
	@scripts/lib.sh
	@helm init -c > /dev/null
	@mkdir -p build

install-helm:
	-kubectl create serviceaccount --namespace kube-system tiller
	-kubectl create clusterrolebinding tiller-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
	-helm init --wait --service-account tiller

delete-local:
	$(MAKE) -C $(CHART) $@

install-local: install-helm delete-local
	$(MAKE) -C $(CHART) $@

install-local-latest: docs/$(RELEASE_LATEST).tgz install-helm delete-local
	$(MAKE) -C $(CHART) $@

# always run these steps if in dependencies:
.PHONY: all build latest release clean lint test minikube-test init install-helm \
	delete-local install-local install-local-latest
