#!/usr/bin/env bash

set -exu

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# Install helm chart
echo "Installing ES from helm charts"
cd $script_dir/../../
CHART=enterprise-suite make install-local-latest

# run our tests
echo "Running tests"
cd $script_dir/../tests
./smoketest
