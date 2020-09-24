#!/usr/bin/env bash

set -exu

KUBERNETES_VERSION="v1.15.2"
MINIKUBE_VERSION="v1.12.1"

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
sudo -E minikube start --vm-driver=none --kubernetes-version ${KUBERNETES_VERSION} --cpus 2 --memory 2000
sudo -E minikube addons enable ingress
sudo chown -R $USER $HOME/.minikube
sudo chgrp -R $USER $HOME/.minikube

# this for loop waits until kubectl can access the api server that Minikube has created
set +e
for i in {1..60}; do # timeout after 2 minutes
    echo "Waiting for minikube to come up..."
    kubectl get po
    res=$?
    if [ ${res} -eq 0 ]; then
        break
    fi
    sleep 2
done
exit ${res}
