#!/usr/bin/env bash

# patch MINIKUBE_CONFIG file with real minikube ip
# Note: minikube ip in travis-ci is different due to it is not running in virtual box

MINIKUBE_CONFIG=cypress/config/minikube.json

sed -i'.orig' -e "s/192\.168\.99\.100/`minikube ip`/g" $MINIKUBE_CONFIG

echo "dump $MINIKUBE_CONFIG"
cat $MINIKUBE_CONFIG
