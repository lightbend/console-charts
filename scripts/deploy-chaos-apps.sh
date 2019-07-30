#!/usr/bin/env bash

set -eu
namespace=$1

get_file() {
    local url=$1
    local file=$2
    curl -sf -H "Authorization: Bearer $GITHUB_TOKEN" "$url" > "$file"
}

mkdir -p build
get_file https://raw.githubusercontent.com/lightbend/console-chaos-apps/master/akka-http/chaos-akka-http.yaml build/chaos-akka-http.yaml
get_file https://raw.githubusercontent.com/lightbend/console-chaos-apps/master/akka/chaos-akka.yaml build/chaos-akka.yaml

# Remove old deployment first to force a redeploy
kubectl delete -n "$namespace" --ignore-not-found deployment -l app.kubernetes.io/part-of=chaos-apps
kubectl apply -n "$namespace" -f build/
