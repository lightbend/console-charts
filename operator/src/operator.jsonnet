// This file generates `manifests/operator.yaml`.

local kubecfg = import "kubecfg.libsonnet";
local kube = import "kube-libsonnet/kube.libsonnet";

local defaultNamespace = "lightbend";
local addNamespace(key, obj) = obj { metadata+: { namespace: defaultNamespace } };

local operatorManifests = {
  'operator.yaml': kubecfg.parseYaml(importstr '../build/console-operator/deploy/operator.yaml')[0],
  'service_account.yaml': kubecfg.parseYaml(importstr '../build/console-operator/deploy/service_account.yaml')[0],

  local tweakRoleRules(idx, rule) = (
    if idx == 0 then rule { resources+: ["serviceaccounts"] } else rule
  ),
  local role = kubecfg.parseYaml(importstr '../build/console-operator/deploy/role.yaml')[0],
  'role.yaml': role { rules: std.mapWithIndex(tweakRoleRules, role.rules) },
  'role_binding.yaml': kubecfg.parseYaml(importstr '../build/console-operator/deploy/role_binding.yaml')[0],

  'cluster_role.yaml': kube.ClusterRole("console-operator") {
    rules: [
      {
        apiGroups: ["rbac.authorization.k8s.io"],
        resources: ["clusterroles", "clusterrolebindings"],
        verbs: ["*"],
      }
    ]
  },

  'cluster_role_binding.yaml': kube.ClusterRoleBinding("console-operator") {
    subjects_+: [ addNamespace("tmp", $["service_account.yaml"]) ],
    roleRef_: $["cluster_role.yaml"],
  },
};

std.mapWithKey(addNamespace, operatorManifests)
