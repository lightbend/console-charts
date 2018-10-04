#!/bin/bash

#set -x

set -eu

# Publish the helm charts to GCS.  This is primarily intended to be run by Travis but is
# usable standalone (assuming you have docker installed and are running from the helm-charts directory).

# Use caution with the -d flag.  If you happen to run this against the production repo but from a local
# helm-charts dir with several tarballs missing for some reason, those tarballs will be removed from the
# production repo.  Deletion is disabled by default.

# This implementation inspired by https://medium.com/google-cloud/google-cloud-sdk-dockerfile-861a0399bbbb

usage () {
    echo "${0##*/} [-d] [-n] | [-h]"
    echo "  -d   Delete objects in repo that are not in docs dir.  Default is false."
    echo "  -n   Dry run.  Show what would be published.  Default is false"
    echo "  -h   Print this help."
    exit 1
}

# GCS project
GCS_PROJECT=es-repo
# GCS bucket within project
: "${GCS_BUCKET:=lightbend-helm-charts}"

# HELM_DIR is helm-charts directory.  Defaults to cwd but can be overridden.
: "${HELM_DIR:=$(pwd)}"

# We'll pull a particular version of the gcloud sdk so things are consistent across builds.
# 'latest' is also an option.
# cf. https://hub.docker.com/r/google/cloud-sdk/
: "${CLOUD_SDK_VERSION:=216.0.0}"
CLOUD_SDK_IMAGE="google/cloud-sdk:${CLOUD_SDK_VERSION}"


## defaults
# Don't delete by default
RSYNC_DELETE=
# Run for real by default
RSYNC_DRY_RUN=

while getopts dnh option
do
    case "$option"
    in
        d) RSYNC_DELETE="-d" ;;
        n) RSYNC_DRY_RUN="-n" ;;
        h) usage ;;
        *) usage ;;
    esac
done

GCLOUD_CONFIG_CID=

cleanup() {
    if [ -n "${GCLOUD_CONFIG_CID}" ] ; then
        docker rm ${GCLOUD_CONFIG_CID}
    fi
}

# Make sure we delete the gcloud-config container
trap cleanup 0

docker pull google/cloud-sdk:${CLOUD_SDK_VERSION}

# Use the decrypted service account credentials to authenticate the command line tool and setup
# volume mounts so credentials and tarballs are accessible to docker
GCLOUD_CONFIG_CID=$( docker run -t -d \
     -v "/tmp/resources:/resources" \
     -v "${HELM_DIR}/docs:/docs" \
     --name gcloud-config ${CLOUD_SDK_IMAGE} \
     gcloud auth activate-service-account --key-file /resources/es-repo-7c1fefe17951.json \
     )

docker run --rm -ti --volumes-from gcloud-config ${CLOUD_SDK_IMAGE} \
    gcloud config set project ${GCS_PROJECT}

# Copy over all tarballs.  Don't include all*.yaml or .nojekyll files.
# Optionally use '-d' and '-n' flags.
docker run --rm -ti --volumes-from gcloud-config ${CLOUD_SDK_IMAGE} \
    gsutil -m rsync ${RSYNC_DELETE} ${RSYNC_DRY_RUN} -c -x "all.*\.yaml|\.nojekyll" /docs gs://${GCS_BUCKET}

# By default GCS sets the file type for index.yaml to 'application/octet-stream'.  Move to 'text/yaml'.
# Note that this increments the Metageneration number for this file in GCS.  (In case you were wondering
# why it seems to be 2 all the time...)
if [ -f "${HELM_DIR}/docs/index.yaml" ] ; then  # should always be there but test anyway...
	docker run --rm -ti --volumes-from gcloud-config ${CLOUD_SDK_IMAGE} \
		gsutil setmeta -h 'Content-Type: text/yaml' gs://${GCS_BUCKET}/index.yaml
fi
