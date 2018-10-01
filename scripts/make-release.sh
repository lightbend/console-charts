#!/usr/bin/env bash

#set -x

set -eu
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
. $script_dir/lib.sh

# script
if [ "$#" -lt 1 ]; then
    echo "$0 <chart-name> [version]"
    echo "leave version blank to use version in Chart.yaml (which is then auto incremented)"
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
if [ -z "$version" ]; then
    version=$(yq r Chart.yaml version)
fi
echo "using version: $version"

# Check we haven't already tagged with this version.
git_tag=$chart-$version
if git rev-parse $git_tag &> /dev/null; then
    echo "$git_tag already exists, check 'git tag'"
    exit 1
fi

echo "Building release"
cd $make_dir
CHARTS=($chart)
if [ "$chart" = "enterprise-suite" ] ; then
    # We build both enterprise-suite and enterprise-suite-latest at the same time...
    CHARTS+=" ${chart}-latest"
fi
make -B CHARTS="${CHARTS[@]}"
git add docs

git commit -m "Release $git_tag"
git tag -a $git_tag -m "Release $git_tag"
echo Tagged release with $git_tag

# Update version for next build
cd $chart_dir
semver=(${version//./ })
((semver[2]++))
next_version="${semver[0]}.${semver[1]}.${semver[2]}"
echo
echo "setting next version to $next_version"
yq w -i Chart.yaml version $next_version
git add Chart.yaml
git commit -m "Incremented version for next release to $next_version"

echo
echo When ready, do a 'git push --follow-tags' to finish the release.
