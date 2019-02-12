#!/bin/bash

cd console-charts
minikube start --vm-driver=none
minikube addons enable ingress
GO111MODULE=on make -C enterprise-suite install-helm gotests-minikube