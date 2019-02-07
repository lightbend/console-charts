package lbc

import (
	"time"

	"github.com/lightbend/gotests/args"
	"github.com/lightbend/gotests/util"
)

const localChartPath = "../../../."
const lbcPath = "../../../scripts/lbc.py"

func Install(namespace string, additionalArgs ...string) error {
	defaultArgs := []string{"install", "--local-chart", localChartPath,
		"--namespace", namespace,
		"--set prometheusDomain=console-backend-e2e.io",
		"--wait", "--", "--timeout 520"}
	fullArgs := append(defaultArgs, additionalArgs...)
	cmd := util.Cmd(lbcPath, fullArgs...)
	if args.TillerNamespace != "" {
		cmd = cmd.Env("TILLER_NAMESPACE", args.TillerNamespace)
	}
	if _, err := cmd.Timeout(time.Minute * 6).Run(); err != nil {
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
