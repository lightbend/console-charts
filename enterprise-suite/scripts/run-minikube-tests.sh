#!/usr/bin/env bash

set -exu

TARGET=$1

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# setup
echo "Installing ES from helm charts, $( basename $( pwd ) )"

make build
make install-local

# run tests
echo "Running $TARGET"
if [ "$TARGET" == "smoketest" ]; then
  cd $script_dir/../tests
  ./smoketest
elif [ "$TARGET" == "browser_e2e" ]; then
  cd $script_dir
  ./run-browser-e2e-tests.sh 
fi
