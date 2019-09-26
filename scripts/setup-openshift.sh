#!/usr/bin/env bash

set -eu

token=$1
project=$2

version="v3.11.0"
full_version="${version}-0cbc58b-linux-64bit"
tarball="openshift-origin-client-tools-${full_version}.tar.gz"
curl -OL https://github.com/openshift/origin/releases/download/${version}/${tarball}
echo "4b0f07428ba854174c58d2e38287e5402964c9a9355f6c359d1242efd0990da3  ${tarball}" | sha256sum -c
tar --strip 1 -xvzf ${tarball} && chmod +x oc && sudo cp oc /usr/local/bin/ && rm oc
oc login https://centralpark2.lightbend.com --token="$token"
oc status
oc project "$project"
oc get pods
echo "Logged in to Openshift cluster"
