#!lib/bats/bats

# Unit tests for install-es.sh.

load 'lib/bats-support/load'
load 'lib/bats-assert/load'

export DRY_RUN=true
install_es=$BATS_TEST_DIRNAME/../scripts/install-es.sh

function setup {
    unset LIGHTBEND_COMMERCIAL_CREDENTIALS
    export LIGHTBEND_COMMERCIAL_USERNAME="myuser"
    export LIGHTBEND_COMMERCIAL_PASSWORD="mypass"
    export ES_STUB_CHART_STATUS="1"
    export ES_HELM_NAME="myhelmname"
}

@test "loads commercial credentials from file" {
    unset LIGHTBEND_COMMERCIAL_USERNAME
    unset LIGHTBEND_COMMERCIAL_PASSWORD
    LIGHTBEND_COMMERCIAL_CREDENTIALS="$BATS_TEST_DIRNAME/testdata/test_credentials" \
        run $install_es
    assert_output --regexp '.*helm install.*--values [^ ]*creds\..*'
}

@test "loads commercial credentials from file with no trailing newline" {
    unset LIGHTBEND_COMMERCIAL_USERNAME
    unset LIGHTBEND_COMMERCIAL_PASSWORD
    LIGHTBEND_COMMERCIAL_CREDENTIALS="$BATS_TEST_DIRNAME/testdata/test_credentials_nonl" \
        run $install_es
    assert_output --regexp '.*helm install.*--values [^ ]*creds\..*'
}

@test "loads commercial credentials from env vars" {
    run $install_es
    assert_output --regexp '.*helm install.*--values [^ ]*creds\..*'
}

@test "adds and updates helm repo if using a published chart" {
    run $install_es
    assert_output --partial "helm repo add es-repo https://repo.lightbend.com/helm-charts"
    assert_output --partial "helm repo update"
}

@test "allows usage of a local chart" {
    ES_LOCAL_CHART="my-local-chart.tgz" \
        run $install_es
    assert_output --partial "helm install my-local-chart.tgz"
    refute_output --partial "helm repo"
}

@test "helm install command" {
    run $install_es
    assert_output --regexp '.*helm install es-repo/enterprise-suite --name myhelmname --namespace lightbend --values [^ ]*creds\..*'
}

@test "helm upgrade command if chart exists" {
    ES_STUB_CHART_STATUS="0" \
        run $install_es
    assert_output --regexp '.*helm upgrade myhelmname es-repo/enterprise-suite --values [^ ]*creds\..*'
}

@test "export yaml files with '--version blah'" {
    ES_EXPORT_YAML=true \
        run $install_es --version v10.0.20 --set minikube=true,podUID=100001
	assert_output --regexp 'helm fetch .*--version v10.0.20 es-repo/enterprise-suite'
    refute_output --regexp "helm fetch [^\n]*--set minikube=true"

	assert_output --regexp 'helm template .*--set minikube=true,podUID=100001 .*/enterprise-suite.*tgz'
    refute_output --regexp 'helm template .*--version v10.0.20'
}

@test "export yaml files with '--version=blah'" {
    ES_EXPORT_YAML=true \
        run $install_es --version=v10.0.20 --set minikube=true,podUID=100001
	assert_output --regexp 'helm fetch .*--version=v10.0.20 es-repo/enterprise-suite'
    refute_output --regexp "helm fetch [^\n]*--set minikube=true"

	assert_output --regexp 'helm template .*--set minikube=true,podUID=100001 .*/enterprise-suite.*tgz'
    refute_output --regexp 'helm template .*--version=v10.0.20'
}

@test "can set namespace" {
    ES_NAMESPACE=mycoolnamespace \
        run $install_es
    assert_output --partial "--namespace mycoolnamespace"
}

@test "can pass helm args" {
    run $install_es --version v10.0.20 --set minikube=true,podUID=100001
    assert_output --partial "--version v10.0.20 --set minikube=true,podUID=100001"
}

@test "warns if no version set" {
    run $install_es
    assert_output --partial "warning: --version"
}

@test "doesn't warn if version set" {
    run $install_es --version v5.0.0
    refute_output --partial "warning: --version"
}

@test "force install deletes the existing install first" {
    ES_STUB_CHART_STATUS="0" ES_FORCE_INSTALL="true" \
        run $install_es
    assert_output --partial "helm delete --purge myhelmname" 
    assert_output --partial "helm install"
    refute_output --partial "helm upgrade"
}
