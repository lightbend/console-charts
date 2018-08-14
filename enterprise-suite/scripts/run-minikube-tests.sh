#!/usr/bin/env bash

set -exu

USE_LATEST_IMAGES=${USE_LATEST_IMAGES:-false}

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# setup
echo "Installing ES from helm charts, [USE_LATEST_IMAGES=$USE_LATEST_IMAGES]"
cd $script_dir/../../

make build CHART=enterprise-suite

if [ "$USE_LATEST_IMAGES" == "true" ]; then
    target=install-local-latest
else
    target=install-local
fi
make $target CHART=enterprise-suite

# run tests
echo "Running tests"
cd $script_dir/../tests
./smoketest
