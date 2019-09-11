
set -eu
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

. "$script_dir"/vars.sh

echo "Building operator image ${full_docker_name}..."

cd "$script_dir"/..
if [[ "$VERSION" == "latest" ]]; then
    echo "Using local helm chart..."
    new_operator_args=--helm-chart="$script_dir"/../../enterprise-suite
else
    echo "Using published ${CONSOLE_TAG} helm chart..."
    new_operator_args=--helm-chart-repo=https://repo.lightbend.com/helm-charts \
        --helm-chart=enterprise-suite --helm-chart-version="${CONSOLE_TAG}"
fi

rm -rf build && mkdir build

cd build
operator-sdk new console-operator --type=helm --kind=Console --api-version=console.lightbend.com/v1alpha1 \
    ${new_operator_args}

cd console-operator
operator-sdk build "${full_docker_name}"

# Create final manifests folder
cd "$script_dir/.."
rm -rf manifests && mkdir manifests
cp -r build/console-operator/deploy/* manifests/

find manifests/
echo "Done creating operator and manifests."
