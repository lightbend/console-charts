#!/usr/bin/env bash

set -e

version="v3.10.0"
full_version="${version}-dd10d17-linux-64bit"
tarball="openshift-origin-client-tools-${full_version}.tar.gz"
curl -OL https://github.com/openshift/origin/releases/download/${version}/${tarball}
echo "0f54235127884309d19b23e8e64e347f783efd6b5a94b49bfc4d0bf472efb5b8  ${tarball}" | sha256sum -c
tar --strip 1 -xvzf ${tarball} && chmod +x oc && sudo cp oc /usr/local/bin/ && rm oc
oc login https://centralpark.lightbend.com --token=$1
oc status
oc project
oc get pods
echo "Logged in to Openshift cluster"
