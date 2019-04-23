package lbc

import (
	"strings"
	"time"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/args"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/util"
)

const localChartPath = "../../../."
const lbcPath = "../../../scripts/lbc.py"

func Install(lbcArgs, helmArgs []string) error {
	cmdArgs := []string{lbcPath, "install", "--local-chart", localChartPath,
		"--namespace", args.ConsoleNamespace,
		"--set prometheusDomain=console-backend-e2e.io"}
	cmdArgs = append(cmdArgs, lbcArgs...)
	cmdArgs = append(cmdArgs, "--", "--wait", "--timeout", "110")
	cmdArgs = append(cmdArgs, helmArgs...)
	cmd := util.Cmd("/bin/bash", "-c", strings.Join(cmdArgs, " "))
	if args.TillerNamespace != "" {
		cmd = cmd.Env("TILLER_NAMESPACE", args.TillerNamespace)
	}

	if err := cmd.Timeout(time.Minute * 2).Run(); err != nil {
		logDebugInfo(args.ConsoleNamespace)
		return err
	}

	return nil
}

func logDebugInfo(namespace string) {
	util.LogG("\n********************************************\nInstall failed, printing debug information:\n\n")

	debugCmds := []string{
		"get events --sort-by=.metadata.resourceVersion",
		"logs -lapp.kubernetes.io/name=lightbend-console,app.kubernetes.io/component=es-console --all-containers",
		"logs -lapp.kubernetes.io/name=lightbend-console,app.kubernetes.io/component=es-console --all-containers -p",
		"logs -lapp.kubernetes.io/name=lightbend-console,app.kubernetes.io/component=prometheus --all-containers",
		"logs -lapp.kubernetes.io/name=lightbend-console,app.kubernetes.io/component=prometheus --all-containers -p",
	}

	for _, cmd := range debugCmds {
		kubectlArgs := []string{"-n", namespace}
		kubectlArgs = append(kubectlArgs, strings.Split(cmd, " ")...)
		if err := util.Cmd("kubectl", kubectlArgs...).PrintCommand().Run(); err != nil {
			util.LogG("Could not gather debug info: %v\n", err)
		}
		util.LogG("\n")
	}
}

func Verify(namespace, tillerNamespace string) error {
	cmd := util.Cmd(lbcPath, "verify", "--namespace", namespace)
	if tillerNamespace != "" {
		cmd = cmd.Env("TILLER_NAMESPACE", tillerNamespace)
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
