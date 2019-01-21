#!/bin/bash

minikube start --cpus=3 --memory=6000
kubectl create serviceaccount --namespace kube-system tiller
kubectl create clusterrolebinding kube-system:tiller --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
helm init --wait --service-account tiller --tiller-namespace=kube-system