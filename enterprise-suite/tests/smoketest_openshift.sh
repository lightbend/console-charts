#!/usr/bin/env bash

# Runs select smoketests on the Openshift cluster, ensuring clean up when done.

(return 2>/dev/null) || set -ux

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
cd $script_dir

source ./smoketest_ci.sh

function test_context() {
    echo openshift
}

function diagnostics() {
    # diagnostics
    oc project
    oc get pod
}

function setup() {
	${script_dir}/../scripts/lbc.py install --namespace=${NAMESPACE} --local-chart=${script_dir}/.. -- \
		--set usePersistentVolumes=true,defaultStorageClass=gp2,prometheusDomain=console-backend-e2e.io --wait
	oc expose service/console-server --namespace=${NAMESPACE}
}

function cleanup() {
    # clean up
    helm del --purge enterprise-suite
    oc delete pvc --all
    oc delete route console-server
}

# Only run main if not sourced.
(return 2>/dev/null) || SMOKE_TESTS="smoke_prometheus" main
