#!/usr/bin/env bash

set -eu
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

. "$script_dir/vars.sh"

if [[ "$VERSION" == "latest" ]]; then
    echo "VERSION is not set, cannot release"
    exit 1
fi

echo "Pushing to docker registry ${docker_registry}..."
docker login -u ${DOCKER_USERNAME} -p ${DOCKER_PASSWORD} ${docker_registry}
docker push ${full_docker_name}
latest="${docker_registry}/${image_name}:latest"
docker tag ${full_docker_name} ${latest}
docker push ${latest}
