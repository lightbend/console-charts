#!/usr/bin/env bash

set -eu

curl -X POST "https://circleci.com/api/v1.1/project/github/lightbend/build-helm-charts/build?circle-token=${CIRCLE_CI_TOKEN}&branch=master"
