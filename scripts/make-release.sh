#!/usr/bin/env bash

set -eu
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
. $script_dir/lib.sh

# script
if [ "$#" -lt 1 ]; then
    echo "$0 <chart-name> [version]"
    echo "leave version blank to auto increment"
    exit 1
fi

chart=$1
version=${2:-}
make_dir=$script_dir/..
chart_dir=$make_dir/$chart

# preflight
if [ ! -z "$(git status -uno --porcelain)" ]; then
    echo "git checkout is not clean, please commit before releasing"
    exit 1
fi

if [ ! -e "$chart/Chart.yaml" ]; then
    echo "$chart/Chart.yaml does not exist"
    exit 1
fi

# release
echo "=== Releasing $chart"
cd $chart_dir
current_version=$(yq r Chart.yaml version)
echo "current version: $current_version"

if [ -z "$version" ]; then
    semver=(${current_version//./ })
    ((semver[2]++))
    version="${semver[0]}.${semver[1]}.${semver[2]}"
fi
echo "setting version to $version"

# Check we haven't already tagged with this version.
git_tag=$chart-$version
if git rev-parse $git_tag &> /dev/null; then
    echo "$git_tag already exists, check 'git tag'"
    exit 1
fi

yq w -i Chart.yaml version $version
git add Chart.yaml

echo "Building release"
cd $make_dir
make -B
git add docs

git commit -m "Release $git_tag"
git tag -a $git_tag -m "Release $git_tag"

echo Tagged commit with $git_tag
echo
echo When ready, do a 'git push --follow-tags' to finish the release.
