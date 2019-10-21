#!/usr/bin/env bash

set -exu

subset=${1:-all}

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

npm run e2e:demo-app-setup
npm install
npm run e2e:patch-minikube-ip
npm run e2e:wait-es-services
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
