#!/usr/bin/env bash

set -eu
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# script
rel_dir=$( dirname $script_dir )
chart_dir=$rel_dir/$1

cd $chart_dir
yq r Chart.yaml version
