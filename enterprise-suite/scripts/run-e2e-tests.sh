#!/usr/bin/env bash

set -exu

subset=$1

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# setup
echo "Installing ES from helm charts, $( basename $( pwd ) )"

NAMESPACE=lightbend
TILLER_NAMESPACE=lightbend
source ${script_dir}/../tests/setup_minikube.sh
setup

# run tests
echo "Running tests"
cd $script_dir/../tests/e2e

# run the e2e test
set +e

npm install
if [[ "$?" != "0" ]]; then
    diagnostics
    exit 1
fi

npm run e2e:demo-app-setup
if [[ "$?" != "0" ]]; then
    diagnostics
    exit 1
fi

npm run e2e:patch-minikube-ip
if [[ "$?" != "0" ]]; then
    diagnostics
    exit 1
fi

npm run e2e:wait-es-services
if [[ "$?" != "0" ]]; then
    diagnostics
    exit 1
fi

if [[ "$subset" == "all" ]]; then
    npm run e2e:travis-prs
elif [[ "$subset" == "1" ]]; then
    npm run e2e:travis-prs-subset1
elif [[ "$subset" == "2" ]]; then
    npm run e2e:travis-prs-subset2
else
    echo "wrong parameter $subset for run-e2e-tests.sh"
    exit 1
fi
if [[ "$?" != "0" ]]; then
    diagnostics
    exit 1
fi
set -e
