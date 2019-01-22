package lbc

import (
	"time"

	"github.com/lightbend/go_tests/args"
	"github.com/lightbend/go_tests/util"
)

const localChartPath = "../../../."
const lbcPath = "../../../scripts/lbc.py"

func Install(namespace string) error {
	cmd := util.Cmd(lbcPath, "install", "--local-chart", localChartPath,
		"--namespace", namespace,
		"--set exposeServices=NodePort",
		"--set prometheusDomain=console-backend-e2e.io",
		"--wait")
	if args.TillerNamespace != "" {
		cmd = cmd.Env("TILLER_NAMESPACE", args.TillerNamespace)
	}
	if _, err := cmd.Timeout(time.Minute * 4).Run(); err != nil {
		return err
	}

	return nil
}

func Verify(namespace string) error {
	cmd := util.Cmd(lbcPath, "verify", "--namespace", namespace)
	if args.TillerNamespace != "" {
		cmd = cmd.Env("TILLER_NAMESPACE", args.TillerNamespace)
	}
	if _, err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func Uninstall() error {
	cmd := util.Cmd(lbcPath, "uninstall")
	if args.TillerNamespace != "" {
		cmd = cmd.Env("TILLER_NAMESPACE", args.TillerNamespace)
	}
	if _, err := cmd.Timeout(time.Minute).Run(); err != nil {
		return err
	}

	return nil
}
