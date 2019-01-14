package lbc

import (
	"time"

	"github.com/lightbend/console_test/util"
)

const localChartPath = "../."
const lbcPath = "../scripts/lbc.py"

func Install(namespace string) error {
	cmd := util.Cmd(lbcPath, "install", "--local-chart", localChartPath,
		"--namespace", namespace, "--set exposeServices=NodePort", "--wait")
	if _, err := cmd.PrintOutput().Timeout(time.Minute * 4).Run(); err != nil {
		return err
	}

	return nil
}

func Verify(namespace string) error {
	if _, err := util.Cmd(lbcPath, "verify", "--namespace", namespace).Run(); err != nil {
		return err
	}

	return nil
}

func Uninstall() error {
	cmd := util.Cmd(lbcPath, "uninstall")
	if _, err := cmd.PrintOutput().Timeout(time.Minute).Run(); err != nil {
		return err
	}

	return nil
}
