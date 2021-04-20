#!/usr/bin/env bash

# Installs helm & console into minikube cluster

(return 2>/dev/null) || set -ux

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
cd $script_dir

export NAMESPACE=${NAMESPACE:-console-backend-e2e}
export TILLER_NAMESPACE=${TILLER_NAMESPACE:-${NAMESPACE}}

function test_context() {
    echo minikube
}

function setup() {
    kubectl create namespace ${NAMESPACE}
    if [ "${NAMESPACE}" != "${TILLER_NAMESPACE}" ]; then
        kubectl create namespace ${TILLER_NAMESPACE}
    fi
    kubectl create serviceaccount --namespace ${TILLER_NAMESPACE} tiller
    kubectl create clusterrolebinding ${TILLER_NAMESPACE}:tiller --clusterrole=cluster-admin \
        --serviceaccount=${TILLER_NAMESPACE}:tiller
    helm init --debug --wait --service-account tiller --upgrade --tiller-namespace=${TILLER_NAMESPACE}

    kubectl config set-context minikube --namespace=${NAMESPACE}

    ${script_dir}/../scripts/lbc.py install --namespace=${NAMESPACE} --local-chart=${script_dir}/.. -- \
        --set podUID=10001,usePersistentVolumes=true,prometheusDomain=console-backend-e2e.io \
        --set exposeServices=NodePort,esConsoleExposePort=30080 \
        --set esConsoleURL=http://console.test.bogus:30080 \
        ${ES_CONSOLE_VERSION+--set esConsoleVersion=${ES_CONSOLE_VERSION}} \
        --wait
}

function cleanup() {
    # optimization - no need for cleanup on travis, as VM will be torn down
    if [[ "${TRAVIS:-}" == "true" ]]; then return; fi

    helm del --purge enterprise-suite
    kubectl delete namespace ${NAMESPACE}
}
