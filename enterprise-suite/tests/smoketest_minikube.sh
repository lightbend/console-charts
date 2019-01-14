#!/usr/bin/env bash

# Runs select smoketests on a local Minikube cluster, ensuring clean up when done.

(return 2>/dev/null) || set -ux

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
cd $script_dir

source ./smoketest_ci.sh

function test_context() {
    echo minikube
}

function diagnostics() {
    # diagnostics
    minikube status
    kubectl get namespace
    kubectl get pod -n ${NAMESPACE}
    kubectl get events -n ${NAMESPACE}
}

function setup() {
	kubectl create namespace ${NAMESPACE}
	kubectl create namespace ${TILLER_NAMESPACE}
	kubectl create serviceaccount --namespace ${TILLER_NAMESPACE} tiller
	kubectl create clusterrolebinding ${TILLER_NAMESPACE}:tiller --clusterrole=cluster-admin \
	    --serviceaccount=${TILLER_NAMESPACE}:tiller
	helm init --wait --service-account tiller --upgrade --tiller-namespace=${TILLER_NAMESPACE}

	kubectl config set-context minikube --namespace=${NAMESPACE}

	${script_dir}/../scripts/lbc.py install --namespace=${NAMESPACE} --local-chart=${script_dir}/.. -- \
		--set podUID=10001,usePersistentVolumes=true,prometheusDomain=console-backend-e2e.io \
		--set exposeServices=NodePort,esConsoleExposePort=30080 \
		--wait
}

function cleanup() {
    # optimization - no need for cleanup on travis, as VM will be torn down
    if [[ "${TRAVIS:-}" == "true" ]]; then return; fi

    helm del --purge enterprise-suite
    kubectl delete namespace ${NAMESPACE}
}

# Only run main if not sourced.
(return 2>/dev/null) || main
