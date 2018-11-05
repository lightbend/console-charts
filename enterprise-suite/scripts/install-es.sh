#!/usr/bin/env bash

#set -x

set -eu

CREDS=
TDIR=

cleanup() {
    if [ -n "$CREDS" -a -f "$CREDS" ] ; then
        rm -f "$CREDS"
    fi
    if [ -n "$TDIR" -a -d "$TDIR" ] ; then
        rm -rf "$TDIR"
    fi
}

# Make sure we delete the credentials file
trap cleanup 0

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
    echo "$0 [-h] | [HELM_ARGS]"
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
    docvar ES_HELM_NAME "Helm release name"
    docvar ES_FORCE_INSTALL "Set to true to delete an existing install first, instead of upgrading"
    docvar ES_EXPORT_YAML "Export resource yaml to stdout.  Set to 'creds' for credentials, 'console' for everything else.  Does not install."
    docvar DRY_RUN "Set to true to dry run the install script"
    exit 1
}

# Create arg for the helm install/upgrade lines.  $1 is username.  $2 is password.
# This prevents the credentials being written to a log.
function set_credentials_arg() {
    CREDS=$(mktemp -t creds.XXXXXX)
    # write creds to file for use by helm
    printf '%s\n' "imageCredentials:" "   username: $1" "   password: $2" >"$CREDS"
    HELM_CREDENTIALS_VALUES="--values $CREDS"
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
        while IFS='=' read -r key value || [ -n "$key" ]; do
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

    set_credentials_arg "$repo_username" "$repo_password"
}

# Echo command to stderr.  Output of running command also goes to stderr by default.
# If "-o X" flag is first argument, stdout of the command is redirected to X.
function debug() {
    command_redirect_to=2
    if [ "$1" = "-o" ] ; then
        command_redirect_to=$2
        shift
        shift
    fi
    echo "$@" >&2
    if [ "false" == "$DRY_RUN" ]; then
        eval "$@" >&$command_redirect_to
    fi
}

function chart_installed() {
    local name=$1
    if [ -n "${ES_STUB_CHART_STATUS:-}" ]; then
        return "$ES_STUB_CHART_STATUS"
    else
        debug "helm status $name > /dev/null 2>&1"
    fi
}

# User overridable variables.
LIGHTBEND_COMMERCIAL_USERNAME=${LIGHTBEND_COMMERCIAL_USERNAME:-}
LIGHTBEND_COMMERCIAL_PASSWORD=${LIGHTBEND_COMMERCIAL_PASSWORD:-}
LIGHTBEND_COMMERCIAL_CREDENTIALS=${LIGHTBEND_COMMERCIAL_CREDENTIALS:-$HOME/.lightbend/commercial.credentials}
ES_REPO=${ES_REPO:-https://repo.lightbend.com/helm-charts}
ES_CHART=${ES_CHART:-enterprise-suite}
ES_NAMESPACE=${ES_NAMESPACE:-lightbend}
ES_LOCAL_CHART=${ES_LOCAL_CHART:-}
ES_HELM_NAME=${ES_HELM_NAME:-enterprise-suite}
ES_FORCE_INSTALL=${ES_FORCE_INSTALL:-false}
ES_EXPORT_YAML=${ES_EXPORT_YAML:-false}
DRY_RUN=${DRY_RUN:-false}

# Help
if [ "${1-:}" == "-h" ]; then
    usage
fi

# Check version and ES_LOCAL_CHART settings
has_version=true
if [[ ! "$*" =~ "--version" ]] ; then
    has_version=false
    if [ -z "$ES_LOCAL_CHART" ] ; then
        echo "warning: --version has not been set, helm will use the latest available version." \
             "It is recommended to use an explicit version." >&2
    fi
fi

# Setup and install helm chart in the repo
if [ -n "$ES_LOCAL_CHART" ]; then
    # Install from a local chart tarball if ES_LOCAL_CHART is set.
    chart_ref=$ES_LOCAL_CHART
else
    debug helm repo add es-repo "$ES_REPO"
    debug helm repo update
    chart_ref=es-repo/$ES_CHART
fi

if [ "false" != "$ES_EXPORT_YAML" ]; then
    # First grab any --version setting and pull it out of the command line args
    # Note that this modifies $@.
    VERSION_ARG=""
    if [ "true" = "$has_version" ] ; then
        REST=("")
        while [[ $# -gt 0 ]]
        do
            arg=$1
            case $arg in
                # Seems helm will accept "--version=foo" as well as the documented "--version foo"
                --version=*)
                    VERSION_ARG="$arg"
                    shift
                    ;;
                --version)
                    VERSION_ARG="--version $2"
                    shift
                    shift
                    ;;
                *)
                    REST+=("$arg")
                    shift
                    ;;
            esac
        done
        if [ -n "$ES_LOCAL_CHART" ] ; then
            echo "warning: --version ignored when ES_LOCAL_CHART is set" >&2
            VERSION_ARG=""
        fi
        set -- "${REST[@]}" # restore the other positional parameters
    fi

    CREDENTIALS_ARG=""
    if [ "creds" == "$ES_EXPORT_YAML" ]; then
        import_credentials
        # This will generate only the credentials yaml.
        CREDENTIALS_ARG="--execute templates/commercial-credentials.yaml $HELM_CREDENTIALS_VALUES"
        echo "warning: credentials in yaml are not encrypted, only base64 encoded. Handle appropriately." >&2
    fi

    if [ -z "$ES_LOCAL_CHART" ]; then
        TDIR=$(mktemp -d 2>/dev/null || mktemp -d -t 'install-es-tdir')
        debug helm fetch -d $TDIR "$VERSION_ARG" "$chart_ref"
        chart_ref=$TDIR/${ES_CHART}*.tgz
    fi
    debug -o 1 helm template --name "$ES_HELM_NAME" --namespace "$ES_NAMESPACE" \
        "$@" \
        "$CREDENTIALS_ARG" \
        $chart_ref

else
    # Determine if we should upgrade or install.
    should_upgrade=
    if chart_installed "$ES_HELM_NAME" ; then
        if [ "true" == "$ES_FORCE_INSTALL" ]; then
            debug helm delete --purge "$ES_HELM_NAME"
            echo "warning: helm delete does not wait for resources to be removed - if the script fails on install, please re-run it." >&2
            should_upgrade=false
        else
            should_upgrade=true
        fi
    else
        should_upgrade=false
    fi

    import_credentials

    if [ "true" == "$should_upgrade" ]; then
        debug helm upgrade "$ES_HELM_NAME" "$chart_ref" \
            "$HELM_CREDENTIALS_VALUES" \
            "$@"
    else
        debug helm install "$chart_ref" --name "$ES_HELM_NAME" --namespace "$ES_NAMESPACE" \
            "$HELM_CREDENTIALS_VALUES" \
            "$@"
    fi
fi
