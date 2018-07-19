#!/usr/bin/env bash

set -eu
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
. $script_dir/lib.sh

# script
if [ "$#" -ne 2 ]; then
    echo "$0 <chart> <version>"
    exit 1
fi

chart=$1
version=$2

# other dirs
docs_dir=$script_dir/../docs
make_dir=$script_dir/..

# preflight
if [ ! -z $(git status -uno --porcelain) ]; then
    echo "git checkout is not clean, please commit before releasing"
    exit 1
fi

# release
echo "setting $1 version to $2"
cd $chart
echo "current version: $(yq r Chart.yaml version)"
yq w -i Chart.yaml version $version
git add Chart.yaml

echo "building release"
cd $make_dir
make
git add docs
git add resources

git commit -m "Release $version"
git tag $version

echo All done, please check the commit.
echo When ready, do a 'git push' to finish the release.
