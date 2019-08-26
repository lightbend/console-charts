#!/usr/bin/env bash

set -e

operator_name="console-operator"
console_version=${CONSOLE_TAG:-"latest"}
docker_registry="lightbend-docker-commercial-registry.bintray.io"
image_name="enterprise-suite/console-operator"
full_docker_name="${docker_registry}/${image_name}:${console_version}"

echo "Building operator image ${full_docker_name}..."

# Delete existing operator resources if they exist
rm -r ${operator_name} || true

if [[ ! -z ${CONSOLE_TAG} ]]; then
    echo "Using published ${CONSOLE_TAG} helm chart..."
    operator-sdk new ${operator_name} --type=helm --kind=LBConsole \
        --api-version=lightbend.com/v1alpha1 --helm-chart-repo=https://repo.lightbend.com/helm-charts \
        --helm-chart=enterprise-suite --helm-chart-version=${CONSOLE_TAG}
else
    echo "Using local helm chart..."
    operator-sdk new ${operator_name} --type=helm --kind=LBConsole \
            --api-version=lightbend.com/v1alpha1 --helm-chart=.
fi

cd console-operator

operator-sdk build ${full_docker_name}

# OS X sed behaves differently in this case
if [ "$(uname)" == "Darwin" ]; then
    sed -i "" "s#REPLACE_IMAGE#${full_docker_name}#g" deploy/operator.yaml
else
    sed -i "s#REPLACE_IMAGE#${full_docker_name}#g" deploy/operator.yaml
fi

cd ..

# If --verify flag was given, check for git changes. Operator k8s resources should be commited, so 
# no diff will be produced after fresh operator build
if [[ $1 == "--validate-diff" ]]; then
    set +e
    git diff --summary --exit-code
    if [[ $? -ne 0 ]]; then
        echo "Found changes after fresh operator build. Did you run 'make update-operator' for your PR?"
        exit 1
    fi
    set -e
fi

# Only push docker image if this is a tagged build
if [[ ! -z ${CONSOLE_TAG} ]]; then
    echo "Pushing to docker registry ${docker_registry}..."
    docker login -u ${DOCKER_USERNAME} -p ${DOCKER_PASSWORD} ${docker_registry}
    docker push ${full_docker_name}
fi
