#!/usr/bin/env bash

set -e

console_version="1.2.0"
docker_registry="lightbend-docker-commercial-registry.bintray.io"
image_name="enterprise-suite/console-operator"
full_docker_name="${docker_registry}/${image_name}:${console_version}"

operator-sdk new console-operator --type=helm --kind=LBConsole \
    --api-version=lightbend.com/v1alpha1 --helm-chart-repo=https://repo.lightbend.com/helm-charts \
    --helm-chart=enterprise-suite --helm-chart-version=${console_version}

cd console-operator

operator-sdk build ${full_docker_name}

docker login -u ${DOCKER_USERNAME} -p ${DOCKER_PASSWORD} ${docker_registry}
docker push ${full_docker_name}
