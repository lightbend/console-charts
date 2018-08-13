test: lint

lint: lint-helm lint-json lint-promql

lint-helm:
	helm lint .

lint-json:
	find . -name \*.json | xargs -tn 1 jq . >/dev/null

lint-promql:
	scripts/validate-promql.sh

minikube-test:
	scripts/run-minikube-tests.sh

delete-local:
	-helm delete --purge es

install-local:
	helm install ../docs/$(RELEASE).tgz --name=es --namespace=lightbend --debug --wait

install-local-latest:
	helm install ../docs/$(RELEASE_LATEST).tgz --name=es --namespace=lightbend --debug --wait

.PHONY: test lint lint-helm lint-json lint-promql minikube-test \
	delete-local install-local install-local-latest
