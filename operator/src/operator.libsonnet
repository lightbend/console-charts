// This file generates `manifests/operator.yaml`.

local kubecfg = import "kubecfg.libsonnet";
local kube = import "kube-libsonnet/kube.libsonnet";

local defaultNamespace = "lightbend";
local addNamespace(key, obj) = obj { metadata+: { namespace: defaultNamespace } };

local operatorManifests = {
  'console_crd.yaml': kubecfg.parseYaml(importstr '../build/console-operator/deploy/crds/console_v1alpha1_console_crd.yaml')[0],
  'console_cr.yaml': kubecfg.parseYaml(importstr '../build/console-operator/deploy/crds/console_v1alpha1_console_cr.yaml')[0],

  'operator.yaml': kubecfg.parseYaml(importstr '../build/console-operator/deploy/operator.yaml')[0],
  'service_account.yaml': kubecfg.parseYaml(importstr '../build/console-operator/deploy/service_account.yaml')[0],

  local tweakRoleRules(idx, rule) = (
    if idx == 0 then rule { resources+: ["serviceaccounts"] } else rule
  ),
  local role = kubecfg.parseYaml(importstr '../build/console-operator/deploy/role.yaml')[0],
  'role.yaml': role { rules: std.mapWithIndex(tweakRoleRules, role.rules) },
  'role_binding.yaml': kubecfg.parseYaml(importstr '../build/console-operator/deploy/role_binding.yaml')[0] {
    subjects: [{
      kind: "ServiceAccount",
      name: "console-operator",
      namespace: defaultNamespace,
    }]
  },
};

std.mapWithKey(addNamespace, operatorManifests)
