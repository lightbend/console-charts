#!/usr/bin/env bash

set -eu
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
. $script_dir/lib.sh

# script
chart_dir=$1

echo "munging chart in $chart_dir..."
cd $chart_dir

# alter chart files
yq w -i Chart.yaml name enterprise-suite-latest
cp $script_dir/munge-values-to-latest.yaml .
yq w -i -s munge-values-to-latest.yaml values.yaml

# cleanup
rm munge-values-to-latest.yaml
