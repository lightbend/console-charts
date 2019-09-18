#!/usr/bin/env bash

set -eu
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

chart=$1
version=$2

# script
rel_dir=$( dirname ${script_dir} )
chart_dir=${rel_dir}/${chart}

cd $chart_dir
yq w -i Chart.yaml version ${version}
