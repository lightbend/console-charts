package helm

import (
	"time"

	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"

	"github.com/lightbend/console_test/util"
	"github.com/lightbend/console_test/util/kube"
)

const ServiceAccountName = "tiller"

func IsInstalled() bool {
	if _, err := util.Cmd("helm", "version").Run(); err != nil {
		return false
	}

	return true
}

func Install(k8sClient *kubernetes.Clientset, namespace string) error {
	// Create namespace, ignore the error because it might already be created
	_ = kube.CreateNamespace(k8sClient, namespace);

	// Create ServiceAccount
	serviceAccount := &apiv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: ServiceAccountName,
		},
	}
	if _, err := k8sClient.CoreV1().ServiceAccounts(namespace).Create(serviceAccount); err != nil {
		return err
	}

	// Bind cluster-admin to the ServiceAccount
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tiller-admin",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "cluster-admin",
		},
		Subjects: []rbacv1.Subject{
			rbacv1.Subject{
				Kind:      "ServiceAccount",
				Name:      ServiceAccountName,
				Namespace: namespace,
			},
		},
	}
	if _, err := k8sClient.RbacV1().ClusterRoleBindings().Create(clusterRoleBinding); err != nil {
		return err
	}

	// Do `helm init`
	cmd := util.Cmd("helm", "init", "--wait", "--service-account", ServiceAccountName, "--upgrade", "--tiller-namespace", namespace)
	if _, err := cmd.PrintOutput().Timeout(time.Minute).Run(); err != nil {
		return err
	}

	return nil
}
