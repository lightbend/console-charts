#!/bin/bash

set -x
set -eu

# Do not prompt for user input when using any Google Cloud SDK methods.
export CLOUDSDK_CORE_DISABLE_PROMPTS=1

#curl -sSL https://sdk.cloud.google.com | bash > /dev/null
#"${HOME}/google-cloud-sdk/bin/gcloud" --quiet components update
#/usr/lib/google-cloud-sdk/bin/gcloud --quiet components update


# Instructions from https://cloud.google.com/sdk/docs/downloads-apt-get
export CLOUD_SDK_REPO="cloud-sdk-$(lsb_release -c -s)"
echo "deb http://packages.cloud.google.com/apt $CLOUD_SDK_REPO main" | \
    sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list
curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
sudo apt-get update && sudo apt-get install google-cloud-sdk

# This line is critical. We setup the SDK to take precedence in our
# environment over the old SDK that may? already be on the machine.
#. "${HOME}/google-cloud-sdk/path.bash.inc"
#. /usr/lib/google-cloud-sdk/path.bash.inc

gcloud version

# Use the decrypted service account credentials to authenticate the command line tool
gcloud auth activate-service-account --key-file resources/es-repo-7c1fefe17951.json

## These bits from helm-master/.circleci/deploy.sh
#: ${GCLOUD_SERVICE_KEY:?"GCLOUD_SERVICE_KEY environment variable is not set"}
#: ${PROJECT_NAME:?"PROJECT_NAME environment variable is not set"}
PROJECT_NAME=es-repo

#"${HOME}/google-cloud-sdk/bin/gcloud" config set project "${PROJECT_NAME}"
/usr/lib/google-cloud-sdk/bin/gcloud config set project "${PROJECT_NAME}"

#  gsutil rsync -d -n $1 gs://$2
gsutil -m rsync -d -n -x "all.*\.yaml|\.nojekyll" docs gs://marcoderama-test
