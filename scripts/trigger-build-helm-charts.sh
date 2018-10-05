#!/usr/bin/env bash

body='{
"request": {
"branch":"master"
}}'

curl -s -X POST \
   -H "Content-Type: application/json" \
   -H "Accept: application/json" \
   -H "Travis-API-Version: 3" \
   -H "Authorization: token $TRAVIS_CI_TOKEN" \
   -d "$body" \
   https://api.travis-ci.com/repo/lightbend%2Fbuild-helm-charts/requests
