#!/usr/bin/env bash

set -eu
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
cd "$script_dir"

: "${VERSION:=latest}"

api_version="console.lightbend.com/v1alpha1"
docker_registry="lightbend-docker-registry.bintray.io"
image_name="enterprise-suite/console-operator"
full_docker_name="${docker_registry}/${image_name}:${VERSION}"

echo "Building operator image ${full_docker_name}..."

if [[ "$VERSION" == "latest" ]]; then
    echo "Using local helm chart..."
    new_operator_args=--helm-chart="$script_dir"/../enterprise-suite
else
    echo "Using published ${CONSOLE_TAG} helm chart..."
    new_operator_args=--helm-chart-repo=https://repo.lightbend.com/helm-charts \
        --helm-chart=enterprise-suite --helm-chart-version="${CONSOLE_TAG}"
fi

rm -rf build && mkdir build

cd build
operator-sdk new console-operator --type=helm --kind=Console --api-version=${api_version} \
    ${new_operator_args}

cd console-operator
operator-sdk build "${full_docker_name}"

sed -i'' "s#REPLACE_IMAGE#${full_docker_name}#g" deploy/operator.yaml
sed -i'' "/Replace\ this\ with\ the\ built\ image\ name/d" deploy/operator.yaml

# Create final manifests folder
cd "$script_dir"
rm -rf manifests && mkdir manifests
cp -r build/console-operator/deploy/* manifests/
cp -r src/* manifests/

find manifests/
echo "Done creating operator and manifests."

# If --verify flag was given, check for git changes. Operator k8s resources should be commited, so
# no diff will be produced after fresh operator build
# if [[ $1 == "--validate-diff" ]]; then
#     set +e
#     git diff --summary --exit-code
#     if [[ $? -ne 0 ]]; then
#         echo "Found changes after fresh operator build. Did you run 'make update-operator' for your PR?"
#         exit 1
#     fi
#     set -e
# fi

# Only push docker image if this is a tagged build, push both current version tag and "latest" tag
# if [[ ! -z ${CONSOLE_TAG} ]]; then
#     echo "Pushing to docker registry ${docker_registry}..."
#     docker login -u ${DOCKER_USERNAME} -p ${DOCKER_PASSWORD} ${docker_registry}
#     docker push ${full_docker_name}
#     latest="${docker_registry}/${image_name}:latest"
#     docker tag ${full_docker_name} ${latest}
#     docker push ${latest}
# fi
