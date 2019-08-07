#!/usr/bin/env bash

# Downloads a file from console-api repo at specified tag.
# Needs GITHUB_TOKEN env variable with a bearer token value!
# Example usage:
# ./pull-console-api.sh v1.0.13 monitors/default-monitors.json

set -u
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
# shellcheck source=../../scripts/lib.sh
. "$script_dir/../../scripts/lib.sh"

# Expect two arguments
if [ $# -ne 2 ]
then
    echo "Error: expected 2 arguments"
    echo "Usage:"
    echo "$0 <tag> <file>"
    exit 1
fi

tag=$1
file=$2
repo=https://raw.githubusercontent.com/lightbend/console-api

# GITHUB_TOKEN must be defined
if [ -z "${GITHUB_TOKEN:-}" ]; then
    if [ -f "$HOME/.config/hub" ]; then
        GITHUB_TOKEN=$(yq r $HOME/.config/hub '[github.com]\[0].oauth_token')
        1>&2 echo "Found $HOME/.config/hub, will use that for oauth"
    else
        echo "Env variable GITHUB_TOKEN is empty; cannot fetch from private repo!"
        exit 1
    fi
fi

function fetch {
    curl -f -H "Authorization: Bearer $GITHUB_TOKEN" "$repo/$1/$2"
}

# First, pull a file that we know exists from master in order to verify bearer token
fetch "master" "README.md" >& /dev/null
if [ $? -ne 0 ]
then
    echo "Unable to fetch any file from the repo, check if GITHUB_TOKEN is valid"
    exit 1
fi

# Next, pull a file that exists from specified tag to check if tag exists
fetch "$tag" "README.md" >& /dev/null
if [ $? -ne 0 ]
then
    echo "Unable to fetch any file from tag $tag, check if it exists"
    exit 1
fi

# Finally, pull the specified file
fetch "$tag" "$file" 2> /dev/null
if [ $? -ne 0 ]
then
    echo "Unable to fetch file $file from tag $tag, check if it exists"
    exit 1
fi
