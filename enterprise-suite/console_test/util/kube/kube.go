package kube

import (
	"github.com/lightbend/console_test/util"
)

// NOTE: These utilities use kubectl, that means they won't work when running tests from inside the cluster.
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
