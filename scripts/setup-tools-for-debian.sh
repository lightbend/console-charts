#!/usr/bin/env bash

set -eux

sudo apt update

# conntrack: required by k8s >=1.18
sudo apt install -y conntrack

# jq
sudo apt install -y jq

# yq
curl -LO https://github.com/mikefarah/yq/releases/download/2.4.0/yq_linux_amd64
echo "99a01ae32f0704773c72103adb7050ef5c5cad14b517a8612543821ef32d6cc9 yq_linux_amd64" | sha256sum --check
sudo mv yq_linux_amd64 /usr/local/bin/yq
sudo chmod +x /usr/local/bin/yq

# promtool
mkdir -p build
cd build
prom_version=2.9.2
prom_file="prometheus-${prom_version}.linux-amd64.tar.gz"
curl -LO https://github.com/prometheus/prometheus/releases/download/v${prom_version}/${prom_file}
echo "19d29910fd0e51765d47b59b9276df016441ad4c6c48e3b27e5aa9acb5d1da26 $prom_file" | sha256sum --check
tar xzvf ${prom_file}
sudo cp prometheus-${prom_version}.linux-amd64/promtool /usr/local/bin/

# Install helm.
# Keep this at helm 2 so we can test that console will install with older versions.
helm_file="helm-v2.17.0-linux-amd64.tar.gz"
curl -LO https://get.helm.sh/${helm_file}
echo "f3bec3c7c55f6a9eb9e6586b8c503f370af92fe987fcbf741f37707606d70296 ${helm_file}" | sha256sum --check
tar xzvf ${helm_file}
chmod +x linux-amd64/helm
sudo cp linux-amd64/helm /usr/local/bin/
helm init -c --stable-repo-url https://charts.helm.sh/stable

# socat is needed for helm init --wait to work
sudo apt-get install -y socat

# kubectl
KUBERNETES_VERSION="v1.17.17"
curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/${KUBERNETES_VERSION}/bin/linux/amd64/kubectl \
    && chmod +x kubectl && sudo cp kubectl /usr/local/bin/ && rm kubectl

# semver
sudo apt install -y node-semver
