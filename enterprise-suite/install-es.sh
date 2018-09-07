#!/usr/bin/env bash

set -eu

function docvar() {
    envvar=$1
    local detail=$2
    eval echo "$envvar=\$${envvar}"
    if [ -n "$detail" ]; then
        echo -e "\t $detail"
    fi
    echo
}

function usage() {
    echo "$0 [-h]"
    echo
    echo "Environment variables that can be set:"
    echo
    docvar LIGHTBEND_COMMERCIAL_CREDENTIALS "Credentials file in property format with username/password"
    docvar ES_REPO "Helm chart repository"
    docvar ES_CHART "Chart name to install from the repository"
    docvar ES_NAMESPACE "Namespace to install ES-Console into"
    docvar ES_MINIKUBE "Set to true to enable minikube specific options"
    docvar ES_LOCAL_CHART "Set to location of local chart tarball"
    docvar ES_UPGRADE "Set to true to perform a helm upgrade instead of an install"
    docvar ES_VERSION "Set to non-empty to install a specific version"
    exit 1
}

function export_credentials() {
    # Exported variables.
    repo_username=
    repo_password=

    # Determine credentials
    if [[ -n "$LIGHTBEND_COMMERCIAL_USERNAME" && -n "$LIGHTBEND_COMMERCIAL_PASSWORD" ]]; then
        repo_username=$LIGHTBEND_COMMERCIAL_USERNAME
        repo_password=$LIGHTBEND_COMMERCIAL_PASSWORD
    elif [ -e $LIGHTBEND_COMMERCIAL_CREDENTIALS ]; then
        while IFS='=' read -r key value; do
            local trimmed_key=$(echo -e "$key" | tr -d '[:space:]')
            local trimmed_value=$(echo -e "$value" | tr -d '[:space:]')
            if [ "$trimmed_key" == "user" ]; then
                repo_username=$trimmed_value
            fi
            if [ "$trimmed_key" == "password" ]; then
                repo_password=$trimmed_value
            fi
        done < $LIGHTBEND_COMMERCIAL_CREDENTIALS
    fi

    if [[ -z "$repo_username" || -z "$repo_password" ]]; then
        echo "Credentials missing, please check your credentials file"
        echo "LIGHTBEND_COMMERCIAL_CREDENTIALS=${LIGHTBEND_COMMERCIAL_CREDENTIALS}"
        echo
        usage
    fi
}

# User overridable variables.
LIGHTBEND_COMMERCIAL_USERNAME=${LIGHTBEND_COMMERCIAL_USERNAME:-}
LIGHTBEND_COMMERCIAL_PASSWORD=${LIGHTBEND_COMMERCIAL_PASSWORD:-}
LIGHTBEND_COMMERCIAL_CREDENTIALS=${LIGHTBEND_COMMERCIAL_CREDENTIALS:-$HOME/.lightbend/commercial_credentials}
ES_REPO=${ES_REPO:-https://lightbend.github.io/helm-charts}
ES_CHART=${ES_CHART:-enterprise-suite}
ES_NAMESPACE=${ES_NAMESPACE:-lightbend}
ES_MINIKUBE=${ES_MINIKUBE:-false}
ES_LOCAL_CHART=${ES_LOCAL_CHART:-}
ES_UPGRADE=${ES_UPGRADE:-false}
ES_VERSION=${ES_VERSION:-}

# Help
if [ "${1-:}" == "-h" ]; then
    usage
fi

# Get credentials
export_credentials

# Setup and install helm chart
if [ -n "$ES_LOCAL_CHART" ]; then
    # Install from a local chart tarball if ES_LOCAL_CHART is set.
    chart_ref=$ES_LOCAL_CHART
else
    helm repo add es-repo $ES_REPO
    helm repo update
    chart_ref=es-repo/$ES_CHART
fi

if [ -n "$ES_VERSION" ]; then
    chart_version="--version=$ES_VERSION"
else
    chart_version=
fi

if [ "true" == "$ES_UPGRADE" ]; then
    helm upgrade es $chart_ref --debug --wait $chart_version \
        --set minikube=$ES_MINIKUBE,imageCredentials.username=${repo_username},imageCredentials.password=${repo_password}
else
    helm install $chart_ref --name=es --namespace=$ES_NAMESPACE --debug --wait $chart_version \
        --set minikube=$ES_MINIKUBE,imageCredentials.username=${repo_username},imageCredentials.password=${repo_password}
fi
