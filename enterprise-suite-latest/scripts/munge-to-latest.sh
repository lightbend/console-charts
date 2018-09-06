#!/usr/bin/env bash

set -eux
esl_dir="$( cd "$( dirname $( dirname "${BASH_SOURCE[0]}" ) )" >/dev/null && pwd )"

helm_script_dir=$esl_dir/../scripts
. $helm_script_dir/lib.sh

# script
chart_dir=$1

echo "munging chart in $chart_dir..."
cd $chart_dir

# alter chart files
yq w -i Chart.yaml name enterprise-suite-latest
yq w -i -s $esl_dir/values-latest.yaml values.yaml
