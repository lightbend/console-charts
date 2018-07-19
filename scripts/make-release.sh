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
git_tag=$chart-$version

# other dirs
docs_dir=$script_dir/../docs
make_dir=$script_dir/..

# preflight
if [ ! -z "$(git status -uno --porcelain)" ]; then
    echo "git checkout is not clean, please commit before releasing"
    exit 1
fi

if git rev-parse $git_tag; then
    echo "$git_tag already exists, check 'git tag'"
    exit 1
fi

# release
echo "current version: $(yq r Chart.yaml version)"
echo "setting version to $2"
cd $chart
yq w -i Chart.yaml version $version
git add Chart.yaml

echo "building release"
cd $make_dir
make
git add docs
git add resources

git commit -m "Release $git_tag"
git tag $git_tag
echo Tagged commit with $git_tag

echo All done, please check the commit.
echo When ready, do a 'git push --tags' to finish the release.
