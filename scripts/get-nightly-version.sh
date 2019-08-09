#!/usr/bin/env bash

set -eu

most_recent_tag=$(git describe --abbrev=0)
next_version=$(semver -i $most_recent_tag)
nightly_tag="$next_version-nightly.$(date -u +%Y%m%d%H%M)"

echo "$nightly_tag"
