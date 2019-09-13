#!/usr/bin/env bash

set -eu
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

. "$script_dir/vars.sh"

echo "Building operator image ${full_docker_name}..."

# Create operator-sdk project and build image
cd "$script_dir"/..
rm -rf build && mkdir build
cd build
operator-sdk new console-operator --type=helm --kind=Console \
    --api-version=console.lightbend.com/v1alpha1 --helm-chart="$script_dir"/../../enterprise-suite

cd console-operator
operator-sdk build "${full_docker_name}"

# Create final manifests folder
cd "$script_dir/.."
rm -rf manifests && mkdir manifests
cp -r build/console-operator/deploy/* manifests/

find manifests/
echo "Done creating operator and manifests."
