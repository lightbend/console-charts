#!/usr/bin/env bash

# Skeleton file for running backend e2e tests in CI.

set -u
set -x
set +e

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
cd $script_dir

export NAMESPACE=${NAMESPACE:-console-backend-e2e}
export TILLER_NAMESPACE=${TILLER_NAMESPACE:-${NAMESPACE}}

function main() {
    cleanup
    setup
    setup_exitcode=$?

    if [[ "$setup_exitcode" != "0" ]]; then
        echo "error setting up tests"
        diagnostics
        cleanup
        exit ${setup_exitcode}
    fi

    TEST_CONTEXT=$(test_context) SMOKE_TESTS=${SMOKE_TESTS:-} ./smoketest
    tests_exitcode=$?

    cleanup

    exit ${tests_exitcode}
}
