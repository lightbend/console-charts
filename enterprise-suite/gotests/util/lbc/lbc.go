package lbc

import (
	"fmt"
	"strings"
	"time"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/minikube"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/args"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/util"
)

const (
	localChartPath = "../../../."
	Path           = "../../../scripts/lbc.py"
)

type Installer struct {
	UsePersistentVolumes string
	MonitorWarmup        string
	ForceDeletePVCs      bool
	HelmWait             bool
	LocalChart           bool
	FailOnWarnings       bool
	AdditionalLBCArgs    []string
	AdditionalHelmArgs   []string
}

func DefaultInstaller() *Installer {
	return &Installer{
		UsePersistentVolumes: "true",
		MonitorWarmup:        "1s",
		ForceDeletePVCs:      true,
		HelmWait:             true,
		LocalChart:           true,
		FailOnWarnings:       false,
	}
}

func (i *Installer) Install() error {
	cmdArgs := []string{Path, "install",
		"--namespace", args.ConsoleNamespace,
		"--set prometheusDomain=console-backend-e2e.io"}
	if i.LocalChart {
		cmdArgs = append(cmdArgs, "--local-chart", localChartPath)
	}
	if i.ForceDeletePVCs {
		cmdArgs = append(cmdArgs, "--delete-pvcs")
	}
	cmdArgs = append(cmdArgs, i.AdditionalLBCArgs...)
	cmdArgs = append(cmdArgs, "--")

	if i.HelmWait {
		cmdArgs = append(cmdArgs, "--wait", "--timeout", "110")
	}
	cmdArgs = append(cmdArgs, "--set esConsoleURL=http://console.test.bogus:30080")
	if minikube.IsRunning() {
		cmdArgs = append(cmdArgs, "--set exposeServices=NodePort")
	}
	if i.UsePersistentVolumes != "" {
		cmdArgs = append(cmdArgs, "--set usePersistentVolumes="+i.UsePersistentVolumes)
	}
	if i.MonitorWarmup != "" {
		cmdArgs = append(cmdArgs, "--set consoleAPI.defaultMonitorWarmup="+i.MonitorWarmup)
	}
	cmdArgs = append(cmdArgs, i.AdditionalHelmArgs...)

	cmd := util.Cmd("/bin/bash", "-c", strings.Join(cmdArgs, " "))
	if args.TillerNamespace != "" {
		cmd = cmd.Env("TILLER_NAMESPACE", args.TillerNamespace)
	}

	var stderr strings.Builder
	if i.FailOnWarnings {
		cmd.CaptureStderr(&stderr)
	}

	if err := cmd.Timeout(time.Minute * 2).Run(); err != nil {
		logDebugInfo(args.ConsoleNamespace)
		return err
	}

	if i.FailOnWarnings {
		for _, line := range strings.Split(stderr.String(), "\n") {
			if strings.HasPrefix(line, "warning:") {
				return fmt.Errorf("found unexpected warning line: %q", line)
			}
		}
	}

	return nil
}

func logDebugInfo(namespace string) {
	util.LogG("\n********************************************\nInstall failed, printing debug information:\n\n")

	debugCmds := []string{
		"get events --sort-by=.metadata.resourceVersion",
		"logs -lapp.kubernetes.io/name=lightbend-console,app.kubernetes.io/component=console-frontend --all-containers",
		"logs -lapp.kubernetes.io/name=lightbend-console,app.kubernetes.io/component=console-frontend --all-containers -p",
		"logs -lapp.kubernetes.io/name=lightbend-console,app.kubernetes.io/component=console-backend --all-containers",
		"logs -lapp.kubernetes.io/name=lightbend-console,app.kubernetes.io/component=console-backend --all-containers -p",
		"logs -lapp.kubernetes.io/name=lightbend-console,app.kubernetes.io/component=grafana --all-containers",
		"logs -lapp.kubernetes.io/name=lightbend-console,app.kubernetes.io/component=grafana --all-containers -p",
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
	cmd := util.Cmd(Path, "verify", "--namespace", namespace)
	if tillerNamespace != "" {
		cmd = cmd.Env("TILLER_NAMESPACE", tillerNamespace)
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
