#!lib/bats/bats

# Unit tests for install-es.sh.

load 'lib/bats-support/load'
load 'lib/bats-assert/load'

export DRY_RUN=true
install_es=$BATS_TEST_DIRNAME/../install-es.sh

function setup {
    unset LIGHTBEND_COMMERCIAL_CREDENTIALS
    export LIGHTBEND_COMMERCIAL_USERNAME="myuser"
    export LIGHTBEND_COMMERCIAL_PASSWORD="mypass"
}

@test "loads commercial credentials from file" {
    unset LIGHTBEND_COMMERCIAL_USERNAME
    unset LIGHTBEND_COMMERCIAL_PASSWORD
    LIGHTBEND_COMMERCIAL_CREDENTIALS="$BATS_TEST_DIRNAME/testdata/test_credentials" \
        run $install_es
    assert_output --partial "imageCredentials.username=testuser,imageCredentials.password=myreallysecurepassword"
}

@test "loads commercial credentials from env vars" {
    run $install_es
    assert_output --partial "imageCredentials.username=myuser,imageCredentials.password=mypass"
}

@test "adds and updates helm repo if using a published chart" {
    run $install_es
    assert_output --partial "helm repo add es-repo https://lightbend.github.io/helm-charts"
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
    assert_output --partial "helm install es-repo/enterprise-suite --name=es --namespace=lightbend --set imageCredentials.username=myuser,imageCredentials.password=mypass"
}

@test "helm upgrade command" {
    ES_UPGRADE=true \
        run $install_es
    assert_output --partial "helm upgrade es es-repo/enterprise-suite --set imageCredentials.username=myuser,imageCredentials.password=mypass"
}

@test "can set namespace" {
    ES_NAMESPACE=mycoolnamespace \
        run $install_es
    assert_output --partial "--namespace=mycoolnamespace"
}

@test "can pass helm args" {
    run $install_es --version v10.0.20 --set minikube=true,podUID=100001
    assert_output --partial "--version v10.0.20 --set minikube=true,podUID=100001"
}
