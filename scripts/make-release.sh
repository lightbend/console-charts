#!/usr/bin/env bash

#set -x

set -eu
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

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
    version_in_chart=$(yq r Chart.yaml version)
    version_stripped=$(echo $version_in_chart | sed 's/\(.*\)-next/\1/')
    # Strip "-next" suffix if it exists
    if [ "$version_in_chart" != "$version_stripped" ] ; then
        version=$version_stripped
    fi
fi
if [ -n "$version" ]; then
    echo "setting version to $version"
    yq w -i Chart.yaml version $version
    git add Chart.yaml
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
# Increment last number-only component and add '-next' if not already there.
# e.g. 1.2.3            -->  1.2.4-next
#      1.2.3-rc.4       -->  1.2.3-rc.5-next
#      1.2.3-rc.4-next  -->  1.2.3-rc.5-next
cd $chart_dir
# This breaks the raw version into           +prefix (ending with '.')
#                                            |     +number
#                                            |     |       +suffix   an optional string starting with . or -, followed
#                                            |     |       |         by alphanumerics (incl. '-') that are not all numbers.
#                                            |     |       |
next_version=$( echo $version | sed -E -e 's/(.*\.)([0-9]+)([\.-]([[:alnum:]]*[[:alpha:]-]+[[:alnum:]-]*))?/\1 \2 \3/' \
        | awk '
             # Increment number component and put version back together, adding "-next" if not already there.
             { newver=$2+1
               if ($3 == ".next") {
                   printf "%s%d%s\n", $1, newver, $3
               } else {
                   printf "%s%d%s-next\n", $1, newver, $3
               }
             }
         '
     )
echo
echo "Setting next version to $next_version"
yq w -i Chart.yaml version $next_version
git add Chart.yaml
git commit -m "Set version for next release to $next_version"

echo
echo When ready, do a 'git push --follow-tags' to finish the release.
