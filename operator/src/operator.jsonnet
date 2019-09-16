// This file generates `manifests/operator.yaml`.

local kubecfg = import "kubecfg.libsonnet";
local kube = import "kube-libsonnet/kube.libsonnet";

local operatorManifests = {
  'operator.yaml': kubecfg.parseYaml(importstr '../build/console-operator/deploy/operator.yaml'),
  'role.yaml': kubecfg.parseYaml(importstr '../build/console-operator/deploy/role.yaml'),
  'role_binding.yaml': kubecfg.parseYaml(importstr '../build/console-operator/deploy/role_binding.yaml'),
  'service_account.yaml': kubecfg.parseYaml(importstr '../build/console-operator/deploy/service_account.yaml'),

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
    // local sa = $["service_account.yaml"] { metadata: { namespace: "placeholder"  }},
    subjects_+: $["service_account.yaml"],
    roleRef_: $["cluster_role.yaml"],
  },
};

operatorManifests
