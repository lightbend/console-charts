
all: docs/es/all.yaml docs/index.yaml

RELEASE := $(shell awk '$$1 == "version:"{v=$$2} $$1 == "name:"{n=$$2} END {print n "-" v}' enterprise-suite/Chart.yaml )

docs/es/all.yaml: docs/$(RELEASE).tgz
	helm template $< > $@

docs/index.yaml: docs/$(RELEASE).tgz
	helm repo index docs --url https://lightbend.github.io/helm-charts

docs/$(RELEASE).tgz: enterprise-suite/Chart.yaml enterprise-suite/templates/*.yaml
	helm init -c
	helm lint enterprise-suite
	helm package enterprise-suite -d docs

lint:
	helm lint enterprise-suite

install-helm:
	-kubectl create serviceaccount --namespace kube-system tiller
	-kubectl create clusterrolebinding tiller-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
	-helm init --service-account tiller

delete-es:
	-helm delete --purge es

install-local: install-helm delete-es
	helm install docs/$(RELEASE).tgz --name=es --namespace=lightbend --debug

clean:
	rm docs/es/all.yaml docs/index.yaml docs/$(RELEASE).tgz

# always run these steps if in dependencies:
.PHONY: all install-local install-helm delete-es lint
