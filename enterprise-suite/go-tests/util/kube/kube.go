package kube

import (
	"github.com/lightbend/go_tests/util"

	"k8s.io/client-go/kubernetes"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: Following two utilities use kubectl, that means they won't work when running tests from inside the cluster.
// A change to parse yaml and use client-go is needed to make that work.

func ApplyYaml(filepath string, namespace string) error {
	if _, err := util.Cmd("kubectl", "-n", namespace, "apply", "-f", filepath).Run(); err != nil {
		return err
	}

	return nil
}

func DeleteYaml(filepath string, namespace string) error {
	if _, err := util.Cmd("kubectl", "-n", namespace, "delete", "-f", filepath).Run(); err != nil {
		return err
	}

	return nil
}

func CreateNamespace(k8sClient *kubernetes.Clientset, name string) error {
	namespace := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	_, err := k8sClient.CoreV1().Namespaces().Create(namespace)
	return err
}

func IsDeploymentAvailable(k8sClient *kubernetes.Clientset, namespace string, name string) bool {
	// NOTE(mitkus): This might be far from best way to figure out if deployment is available

	dep, err := k8sClient.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return false
	}
	if len(dep.Status.Conditions) > 0 {
		if dep.Status.Conditions[0].Type == "Available" {
			return true
		}
	}
	return false
}
