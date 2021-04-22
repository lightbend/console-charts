#!/usr/bin/env bash

set -eu

vcs=github
org=lightbend
project=console-gui

# See details about CircleCI API:
# https://circleci.com/docs/api/v2/#trigger-a-new-pipeline
# https://circleci.com/docs/api/v2/#section/Authentication/api_key_header
curl -v --request POST \
  --url https://circleci.com/api/v2/project/${vcs}/${org}/${project}/pipeline \
  --header "Circle-Token: ${CIRCLECI_API_KEY}" \
  --header "Content-Type: application/json" \
  --data '{ "branch": "master" }'
