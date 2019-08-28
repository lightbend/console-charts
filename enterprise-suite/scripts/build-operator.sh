#!/usr/bin/env bash

set -e

operator_name="console-operator"
api_version="console.lightbend.com/v1alpha1"
console_version=${CONSOLE_TAG:-"latest"}
docker_registry="lightbend-docker-commercial-registry.bintray.io"
image_name="enterprise-suite/console-operator"
full_docker_name="${docker_registry}/${image_name}:${console_version}"

echo "Building operator image ${full_docker_name}..."

# Preserve these files, they don't get generated by operator-sdk
preserve_files=(
    "deploy/cluster_role.yaml"
    "deploy/cluster_role_binding.yaml"
    # TODO: kustomize
)

# Copy preserved files to a temp dir
preserve_dir=$(mktemp -d)
for f in "${preserve_files[@]}"
do
    dest=$(dirname "${preserve_dir}/${f}")
    mkdir -p ${dest} && cp "${operator_name}/${f}" ${dest}
done

# Delete existing operator resources if they exist, operator-sdk requires that destination
# directory doesn't exist.
rm -r ${operator_name} || true

if [[ ! -z ${CONSOLE_TAG} ]]; then
    echo "Using published ${CONSOLE_TAG} helm chart..."
    operator-sdk new ${operator_name} --type=helm --kind=Console \
        --api-version=${api_version} --helm-chart-repo=https://repo.lightbend.com/helm-charts \
        --helm-chart=enterprise-suite --helm-chart-version=${CONSOLE_TAG}
else
    echo "Using local helm chart..."
    operator-sdk new ${operator_name} --type=helm --kind=Console \
            --api-version=${api_version} --helm-chart=.
fi

# Restore preserved files
for f in "${preserve_files[@]}"
do
    cp "${preserve_dir}/${f}" "${operator_name}/${f}"
done
rm -rf ${preserve_dir}

cd console-operator
operator-sdk build ${full_docker_name}

# OS X sed behaves differently in this case
if [ "$(uname)" == "Darwin" ]; then
    sed -i "" "s#REPLACE_IMAGE#${full_docker_name}#g" deploy/operator.yaml
    sed -i "" "/Replace\ this\ with\ the\ built\ image\ name/d" deploy/operator.yaml
else
    sed -i "s#REPLACE_IMAGE#${full_docker_name}#g" deploy/operator.yaml
    sed -i "/Replace\ this\ with\ the\ built\ image\ name/d" deploy/operator.yaml
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

# Only push docker image if this is a tagged build, push both current version tag and "latest" tag
if [[ ! -z ${CONSOLE_TAG} ]]; then
    echo "Pushing to docker registry ${docker_registry}..."
    docker login -u ${DOCKER_USERNAME} -p ${DOCKER_PASSWORD} ${docker_registry}
    docker push ${full_docker_name}
    latest="${docker_registry}/${image_name}:latest"
    docker tag ${full_docker_name} ${latest}
    docker push ${latest}
fi
