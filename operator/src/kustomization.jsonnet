local k = import "kube-libsonnet/kube.libsonnet";
local o = import "operator.libsonnet";

{
  apiVersion: "kustomize.config.k8s.io/v1beta1",
  kind: "Kustomization",

  namespace: k.objectValues(o)[0].metadata.namespace,

  images: [{
    name: "REPLACE_IMAGE",
    newName: "lightbend-docker-registry.bintray.io/lightbend/console-operator",
    newTag: std.extVar("version"),
  }],

  resources: [f for f in std.objectFields(o) if f != "console_cr.yaml"],
}
