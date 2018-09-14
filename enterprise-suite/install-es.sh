#!/usr/bin/env bash

set -eu

function docvar() {
    envvar=$1
    local detail=$2
    eval echo "$envvar=\$${envvar}"
    if [ -n "$detail" ]; then
        echo -e "\\t $detail"
    fi
    echo
}

function usage() {
    echo "$0 [-h] [HELM_ARGS]"
    echo
    echo "-h: prints help"
    echo "HELM_ARGS: will get passed directly to helm"
    echo
    echo "Environment variables that can be set:"
    echo
    docvar LIGHTBEND_COMMERCIAL_CREDENTIALS "Credentials file in property format with username/password"
    docvar ES_REPO "Helm chart repository"
    docvar ES_CHART "Chart name to install from the repository"
    docvar ES_NAMESPACE "Namespace to install ES-Console into"
    docvar ES_LOCAL_CHART "Set to location of local chart tarball"
    docvar ES_UPGRADE "Set to true to perform a helm upgrade instead of an install"
    docvar DRY_RUN "Set to true to dry run the install script"
    exit 1
}

function import_credentials() {
    # Credential variables to set.
    repo_username=
    repo_password=

    # Determine credentials
    if [[ -n "$LIGHTBEND_COMMERCIAL_USERNAME" && -n "$LIGHTBEND_COMMERCIAL_PASSWORD" ]]; then
        repo_username=$LIGHTBEND_COMMERCIAL_USERNAME
        repo_password=$LIGHTBEND_COMMERCIAL_PASSWORD
    elif [ -e "$LIGHTBEND_COMMERCIAL_CREDENTIALS" ]; then
        while IFS='=' read -r key value; do
            local trimmed_key
            trimmed_key=$(echo -e "$key" | tr -d '[:space:]')
            local trimmed_value
            trimmed_value=$(echo -e "$value" | tr -d '[:space:]')
            if [ "$trimmed_key" == "user" ]; then
                repo_username=$trimmed_value
            elif [ "$trimmed_key" == "password" ]; then
                repo_password=$trimmed_value
            fi
        done < "$LIGHTBEND_COMMERCIAL_CREDENTIALS"
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
ES_LOCAL_CHART=${ES_LOCAL_CHART:-}
ES_UPGRADE=${ES_UPGRADE:-false}
DRY_RUN=${DRY_RUN:-false}

# Help
if [ "${1-:}" == "-h" ]; then
    usage
fi

# Setup dry-run
debug=
if [ "$DRY_RUN" == "true" ]; then
    debug=echo
fi

# Get credentials
import_credentials

# Setup and install helm chart
if [ -n "$ES_LOCAL_CHART" ]; then
    # Install from a local chart tarball if ES_LOCAL_CHART is set.
    chart_ref=$ES_LOCAL_CHART
else
    $debug helm repo add es-repo "$ES_REPO"
    $debug helm repo update
    chart_ref=es-repo/$ES_CHART
fi

if [ "true" == "$ES_UPGRADE" ]; then
    $debug helm upgrade es "$chart_ref" \
        --set imageCredentials.username="$repo_username",imageCredentials.password="$repo_password" \
        $@
else
    $debug helm install "$chart_ref" --name=es --namespace="$ES_NAMESPACE" \
        --set imageCredentials.username="$repo_username",imageCredentials.password="$repo_password" \
        $@
fi
