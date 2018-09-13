#!/bin/bash

set -x
set -eu

# Do not prompt for user input when using any Google Cloud SDK methods.
export CLOUDSDK_CORE_DISABLE_PROMPTS=1

# Credentials for system account
#openssl aes-256-cbc -K $encrypted_f01ffbb90c44_key -iv $encrypted_f01ffbb90c44_iv -in resources/es-repo-7c1fefe17951.json.enc -out resources/es-repo-7c1fefe17951.json -d

# The sdk install script errors if this directory already exists,
# but Travis already created it when we marked it as cached.
if [ ! -d $HOME/google-cloud-sdk/bin ]; then
    rm -rf $HOME/google-cloud-sdk;
    curl -sSL https://sdk.cloud.google.com | bash > /dev/null;
	#. ${HOME}/google-cloud-sdk/path.bash.inc
	echo "PATH=${PATH}"
fi

# This line is critical. We setup the SDK to take precedence in our
# environment over the old SDK that may? already be on the machine.
. $HOME/google-cloud-sdk/path.bash.inc

#- gcloud components update kubectl
gcloud version

## This from https://cloud.google.com/solutions/continuous-delivery-with-travis-ci

#mkdir -p lib

# Here we use the decrypted service account credentials to authenticate the command line tool
gcloud auth activate-service-account --key-file resources/es-repo-7c1fefe17951.json

## These bits from helm-master/.circleci/deploy.sh
#: ${GCLOUD_SERVICE_KEY:?"GCLOUD_SERVICE_KEY environment variable is not set"}
: ${PROJECT_NAME:?"PROJECT_NAME environment variable is not set"}

${HOME}/google-cloud-sdk/bin/gcloud config set project "${PROJECT_NAME}"

