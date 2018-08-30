#!/usr/bin/env bash

set -eux

# yq
sudo add-apt-repository -y ppa:rmescandon/yq
sudo apt update
sudo apt install -y yq

# jq
sudo apt install -y jq

# sponge
sudo apt install -y moreutils

# promtool
mkdir -p build
cd build
prom_version=2.3.2
prom_file="prometheus-${prom_version}.linux-amd64.tar.gz"
curl -LO https://github.com/prometheus/prometheus/releases/download/v${prom_version}/${prom_file}
echo "351931fe9bb252849b7d37183099047fbe6d7b79dcba013fb6ae19cc1bbd8552 $prom_file" | sha256sum --check
tar xzvf $prom_file
sudo cp prometheus-${prom_version}.linux-amd64/promtool /usr/local/bin/

# install helm
curl https://raw.githubusercontent.com/kubernetes/helm/master/scripts/get > get_helm.sh
chmod 777 get_helm.sh
sudo ./get_helm.sh
helm init -c

# socat is needed for helm init --wait to work
sudo apt-get install -y socat
