#!/usr/bin/env bash

set -exu

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# setup
echo "Installing ES from helm charts, $( basename $( pwd ) )"

make build
make install-local

# run tests
echo "Running tests"
cd $script_dir/../tests/e2e

# run the e2e test
sudo apt update
sudo apt -y install libgconf2-4
npm install
npm run e2e:demo-app-setup
npm run e2e:patch-minikube-ip
npm run e2e:wait-es-services
npm run e2e:travis-prs
