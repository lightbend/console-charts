// This file generates `kustomization.yaml`.

local k = import "kube-libsonnet/kube.libsonnet";
local o = import "operator.jsonnet";

{
  apiVersion: "kustomize.config.k8s.io/v1beta1",
  kind: "Kustomization",

  namespace: k.objectValues(o)[0].metadata.namespace,

  images: [{
    name: "REPLACE_IMAGE",
    newName: "lightbend-docker-registry.bintray.io/enterprise-suite/console-operator",
    newTag: std.extVar("version"),
  }],

  resources: ["manifests/console-operator.yaml"],
}
