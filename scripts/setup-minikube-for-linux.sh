#!/usr/bin/env bash

set -exu

# Preflight checks
if ! command -v helm > /dev/null; then
    echo "install helm binary"
    exit 1
fi

# MINIKUBE_VERSION="latest" -- https://github.com/kubernetes/minikube/issues/2704
MINIKUBE_VERSION="latest"

# From https://github.com/kubernetes/minikube#linux-continuous-integration-without-vm-support
curl -Lo minikube https://storage.googleapis.com/minikube/releases/${MINIKUBE_VERSION}/minikube-linux-amd64 && chmod +x minikube && sudo cp minikube /usr/local/bin/ && rm minikube

export MINIKUBE_WANTUPDATENOTIFICATION=false
export MINIKUBE_WANTREPORTERRORPROMPT=false
export MINIKUBE_HOME=$HOME
export CHANGE_MINIKUBE_NONE_USER=true
mkdir -p $HOME/.kube
touch $HOME/.kube/config

export KUBECONFIG=$HOME/.kube/config
sudo -E minikube start --vm-driver=none
sudo -E minikube addons enable ingress

# this for loop waits until kubectl can access the api server that Minikube has created
set +e
for i in {1..150}; do # timeout for 5 minutes
    kubectl get po &> /dev/null
    if [ $? -ne 1 ]; then
        break
    fi
    sleep 2
done
