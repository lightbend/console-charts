#!/bin/bash

oc login https://centralpark.lightbend.com --token=$1
oc status && oc project && oc get pods
oc GO111MODULE=on make -C enterprise-suite gotests-openshift NAMESPACE=console-backend-go-tests