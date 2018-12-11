package helm

import (
	"fmt"

	"github.com/lightbend/console_test/util"
)

const ServiceAccountName = "tiller"

func IsInstalled() bool {
	if err := util.ExecZeroExitCode("helm", "status"); err != nil {
		return false
	}

	return true
}

func Install(namespace string) error {
	if err := util.ExecZeroExitCode("kubectl", "create", "namespace", namespace); err != nil {
		return err
	}
	if err := util.ExecZeroExitCode("kubectl", "create", "serviceaccount", "--namespace", namespace, ServiceAccountName); err != nil {
		return err
	}

	namespacedServiceAccount := fmt.Sprintf("%v:%v", namespace, ServiceAccountName)
	err := util.ExecZeroExitCode("kubectl", "create", "clusterrolebinding",
		namespacedServiceAccount, "--clusterrole=cluster-admin", "--serviceaccount", namespacedServiceAccount)
	if err != nil {
		return err
	}

	return nil
}
