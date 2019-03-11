package lbc

import (
	"fmt"
	"strings"
	"time"

	"github.com/onsi/ginkgo"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/args"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util"
)

const localChartPath = "../../../."
const lbcPath = "../../../scripts/lbc.py"

func Install(namespace string, additionalLbcArgs string, additionalArgs ...string) error {
	defaultArgs := []string{"install", "--local-chart", localChartPath,
		"--namespace", namespace,
		"--set prometheusDomain=console-backend-e2e.io",
		"--wait",
		additionalLbcArgs,
		"--", "--timeout 110"}
	fullArgs := append(defaultArgs, additionalArgs...)
	logDebugInfo(fmt.Sprintf("invoking: lbc.py %s %s", lbcPath, fullArgs))
	cmd := util.Cmd(lbcPath, fullArgs...)
	if args.TillerNamespace != "" {
		cmd = cmd.Env("TILLER_NAMESPACE", args.TillerNamespace)
	}

	if err := cmd.Timeout(time.Minute * 2).Run(); err != nil {
		logDebugInfo(namespace)
		return err
	}

	return nil
}

func logDebugInfo(namespace string) {
	fmt.Fprint(ginkgo.GinkgoWriter, "\n********************************************\nInstall failed, printing debug information:\n\n")

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
			fmt.Fprintf(ginkgo.GinkgoWriter, "Could not gather debug info: %v\n", err)
		}
		fmt.Fprintf(ginkgo.GinkgoWriter, "\n")
	}
}

func Verify(namespace string) error {
	cmd := util.Cmd(lbcPath, "verify", "--namespace", namespace)
	if args.TillerNamespace != "" {
		cmd = cmd.Env("TILLER_NAMESPACE", args.TillerNamespace)
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func Uninstall() error {
	cmd := util.Cmd(lbcPath, "uninstall", "--delete-pvcs")
	if args.TillerNamespace != "" {
		cmd = cmd.Env("TILLER_NAMESPACE", args.TillerNamespace)
	}
	if err := cmd.Timeout(time.Minute).Run(); err != nil {
		return err
	}

	return nil
}
