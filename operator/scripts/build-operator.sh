#!/usr/bin/env bash

set -eu
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

. "$script_dir/vars.sh"

echo "Building operator image ${full_docker_name}..."
echo "Checking if minikube is available - this is required to use operator-sdk for helm charts"
if ! minikube status > /dev/null; then
    echo "minikube is down"
    exit 1
fi
if ! kubectl version > /dev/null; then
    echo "kubernetes inaccessible"
    exit 1
fi

# Create operator-sdk project and build image
cd "$script_dir"/..
rm -rf build && mkdir build
cd build
operator-sdk new console-operator --type=helm --kind=Console \
    --api-version=console.lightbend.com/v1alpha1 --helm-chart="$script_dir"/../../enterprise-suite

cd console-operator
operator-sdk build "${full_docker_name}"

# Create OLM manifests for operatorhub.io

# Create final manifests folder
cd "$script_dir/.."
rm -rf manifests && mkdir manifests
for f in src/*.jsonnet; do
    base=$(basename "$f")
    base=${base%.jsonnet}
    kubecfg -J vendor --ext-str version="$VERSION" show -o yaml src/"$base.jsonnet" > manifests/"$base.yaml"
done

find manifests/
echo "Done creating operator and manifests."
