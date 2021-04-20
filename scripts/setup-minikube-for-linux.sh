#!/usr/bin/env bash

set -exu

KUBERNETES_VERSION="v1.15.2"
MINIKUBE_VERSION="latest"

# Preflight checks
if ! command -v helm > /dev/null; then
    echo "install helm binary"
    exit 1
fi


# From https://github.com/kubernetes/minikube#linux-continuous-integration-without-vm-support
curl -Lo minikube https://storage.googleapis.com/minikube/releases/${MINIKUBE_VERSION}/minikube-linux-amd64 && chmod +x minikube && sudo cp minikube /usr/local/bin/ && rm minikube

export MINIKUBE_WANTUPDATENOTIFICATION=false
export MINIKUBE_WANTREPORTERRORPROMPT=false
export MINIKUBE_HOME=$HOME
export CHANGE_MINIKUBE_NONE_USER=true
mkdir -p $HOME/.kube
touch $HOME/.kube/config

export KUBECONFIG=$HOME/.kube/config
# --addons=[]: Enable addons. see `minikube addons list` for a list of valid addon names.
# --wait=all: wait for and verify all Kubernetes components after starting the cluster.
sudo -E minikube start --vm-driver=none --kubernetes-version ${KUBERNETES_VERSION} --cpus 2 --memory 2000 --addons=ingress --wait=all
sudo chown -R $USER $HOME/.minikube
sudo chgrp -R $USER $HOME/.minikube
