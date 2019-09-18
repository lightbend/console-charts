#!/usr/bin/env bash
#
# Used for whitesource image scans.

set -eu
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

values="$script_dir/../enterprise-suite/values.yaml"
consoleVersion=$(yq r "$values" esConsoleVersion)
monitorVersion=$(yq r "$values" esMonitorVersion)
grafanaVersion=$(yq r "$values" esGrafanaVersion)
promImage=$(yq r "$values" prometheusImage)
promVersion=$(yq r "$values" prometheusVersion)
configMapReloadImage=$(yq r "$values" configMapReloadImage)
configMapReloadVersion=$(yq r "$values" configMapReloadVersion)
kubeStateMetricsImage=$(yq r "$values" kubeStateMetricsImage)
kubeStateMetricsVersion=$(yq r "$values" kubeStateMetricsVersion)
goDnsmasqImage=$(yq r "$values" goDnsmasqImage)
goDnsmasqVersion=$(yq r "$values" goDnsmasqVersion)
alpineImage=$(yq r "$values" alpineImage)
alpineVersion=$(yq r "$values" alpineVersion)
busyboxImage=$(yq r "$values" busyboxImage)
busyboxVersion=$(yq r "$values" busyboxVersion)

bintray=lightbend-docker-commercial-registry.bintray.io
docker login -u "$LIGHTBEND_COMMERCIAL_USERNAME" -p "$LIGHTBEND_COMMERCIAL_PASSWORD" "$bintray"
consoleRepo="$bintray/enterprise-suite"
docker pull "$consoleRepo/es-console:$consoleVersion"
docker pull "$consoleRepo/console-api:$monitorVersion"
docker pull "$consoleRepo/es-grafana:$grafanaVersion"
docker pull "$promImage:$promVersion"
docker pull "$configMapReloadImage:$configMapReloadVersion"
docker pull "$kubeStateMetricsImage:$kubeStateMetricsVersion"
docker pull "$goDnsmasqImage:$goDnsmasqVersion"
docker pull "$alpineImage:$alpineVersion"
docker pull "$busyboxImage:$busyboxVersion"
