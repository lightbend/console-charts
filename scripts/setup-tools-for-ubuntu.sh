#!/usr/bin/env bash

set -eux

# git-lfs repo key, needed for apt update as of 2018/01/07 on travis
sudo apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 762E3157

# yq
sudo add-apt-repository -y ppa:rmescandon/yq
sudo apt update
sudo apt install -y yq

# jq
sudo apt install -y jq

# libgconf2-4
sudo apt install -y libgconf2-4

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
# Keep this at 2.10.0 so we can test that console will install with older versions.
helm_file="helm-v2.10.0-linux-amd64.tar.gz"
curl -LO https://storage.googleapis.com/kubernetes-helm/${helm_file}
echo "0fa2ed4983b1e4a3f90f776d08b88b0c73fd83f305b5b634175cb15e61342ffe ${helm_file}" | sha256sum --check
tar xzvf ${helm_file}
sudo cp linux-amd64/helm /usr/local/bin/
helm init -c

# socat is needed for helm init --wait to work
sudo apt-get install -y socat

# kubectl
curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl \
    && chmod +x kubectl && sudo cp kubectl /usr/local/bin/ && rm kubectl
