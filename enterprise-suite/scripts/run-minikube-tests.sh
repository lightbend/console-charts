#!/usr/bin/env bash

set -exu

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# setup
echo "Installing ES from helm charts, $( basename $( pwd ) )"

make build
make install-local

# run tests
echo "Running tests"
cd $script_dir/../tests
./smoketest
