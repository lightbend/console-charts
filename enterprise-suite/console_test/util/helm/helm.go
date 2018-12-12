package helm

import (
	"fmt"
	"time"

	"github.com/lightbend/console_test/util"
)

const ServiceAccountName = "tiller"

func IsInstalled() bool {
	if _, err := util.Cmd("helm", "status").Run(); err != nil {
		return false
	}

	return true
}

func Install(namespace string) error {
	cmd := util.Cmd("kubectl", "create", "namespace", namespace)
	if _, err := cmd.Run(); err != nil {
		return err
	}

	cmd = util.Cmd("kubectl", "create", "serviceaccount", "--namespace", namespace, ServiceAccountName)
	if _, err := cmd.Run(); err != nil {
		return err
	}

	namespacedServiceAccount := fmt.Sprintf("%v:%v", namespace, ServiceAccountName)
	cmd = util.Cmd("kubectl", "create", "clusterrolebinding",
		namespacedServiceAccount, "--clusterrole=cluster-admin", "--serviceaccount", namespacedServiceAccount)
	if _, err := cmd.Run(); err != nil {
		return err
	}

	cmd = util.Cmd("helm", "init", "--wait", "--service-account", ServiceAccountName, "--upgrade", "--tiller-namespace", namespace)
	if _, err := cmd.Timeout(time.Minute).Run(); err != nil {
		return err
	}

	return nil
}
