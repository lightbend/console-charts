package helm

import (
	"fmt"
	"time"

	"github.com/lightbend/console_test/util"
)

const ServiceAccountName = "tiller"
const TillerNamespace = "kube-system"

func IsInstalled() bool {
	if _, err := util.Cmd("helm", "version").Run(); err != nil {
		return false
	}

	return true
}

func Install() error {
	cmd := util.Cmd("kubectl", "create", "serviceaccount", "--namespace", TillerNamespace, ServiceAccountName)
	if _, err := cmd.PrintOutput().Run(); err != nil {
		return err
	}

	namespacedServiceAccount := fmt.Sprintf("%v:%v", TillerNamespace, ServiceAccountName)
	cmd = util.Cmd("kubectl", "create", "clusterrolebinding",
		namespacedServiceAccount, "--clusterrole=cluster-admin", "--serviceaccount", namespacedServiceAccount)
	if _, err := cmd.PrintOutput().Run(); err != nil {
		return err
	}

	cmd = util.Cmd("helm", "init", "--wait", "--service-account", ServiceAccountName, "--upgrade", "--tiller-namespace", TillerNamespace)
	if _, err := cmd.PrintOutput().Timeout(time.Minute).Run(); err != nil {
		return err
	}

	return nil
}
