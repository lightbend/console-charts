#!/usr/bin/env bash
#
# Used for whitesource image scans.

set -eu
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

values="$script_dir/../enterprise-suite/values.yaml"
consoleVersion=$(yq e .esConsoleVersion "$values")
monitorVersion=$(yq e .esMonitorVersion "$values")
grafanaVersion=$(yq e .esGrafanaVersion "$values")
promImage=$(yq e .prometheusImage "$values")
promVersion=$(yq e .prometheusVersion "$values")
configMapReloadImage=$(yq e .configMapReloadImage "$values")
configMapReloadVersion=$(yq e .configMapReloadVersion "$values")
kubeStateMetricsImage=$(yq e .kubeStateMetricsImage "$values")
kubeStateMetricsVersion=$(yq e .kubeStateMetricsVersion "$values")
goDnsmasqImage=$(yq e .goDnsmasqImage "$values")
goDnsmasqVersion=$(yq e .goDnsmasqVersion "$values")
alpineImage=$(yq e .alpineImage "$values")
alpineVersion=$(yq e .alpineVersion "$values")
busyboxImage=$(yq e .busyboxImage "$values")
busyboxVersion=$(yq e .busyboxVersion "$values")

commercial_registry=commercial-registry.lightbend.com
docker login -u "$LIGHTBEND_COMMERCIAL_USERNAME" -p "$LIGHTBEND_COMMERCIAL_PASSWORD" "$commercial_registry"
consoleRepo="$commercial_registry/enterprise-suite"
docker pull "${consoleRepo}-es-console:$consoleVersion"
docker pull "${consoleRepo}-console-api:$monitorVersion"
docker pull "${consoleRepo}-es-grafana:$grafanaVersion"
docker pull "$promImage:$promVersion"
docker pull "$configMapReloadImage:$configMapReloadVersion"
docker pull "$kubeStateMetricsImage:$kubeStateMetricsVersion"
docker pull "$goDnsmasqImage:$goDnsmasqVersion"
docker pull "$alpineImage:$alpineVersion"
docker pull "$busyboxImage:$busyboxVersion"
