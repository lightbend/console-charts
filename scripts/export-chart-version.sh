#!/usr/bin/env bash

awk '$1 == "version:" {print $2}' ./enterprise-suite/Chart.yaml 

function original() {
set -eu
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
. $script_dir/lib.sh

# script
chart_dir=$1

cd $chart_dir
yq r Chart.yaml version

}
