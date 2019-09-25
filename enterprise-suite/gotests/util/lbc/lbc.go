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
	HelmWait             string
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
		HelmWait:             "180",
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
	cmdArgs = append(cmdArgs, "--wait", "--timeout", i.HelmWait)
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

	if err := cmd.Timeout(time.Minute * 5).Run(); err != nil {
		util.LogDebugInfo()
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

func (i *Installer) Uninstall() error {
	cmdArgs := []string{Path, "uninstall", "--namespace", args.ConsoleNamespace}
	cmd := util.Cmd("/bin/bash", "-c", strings.Join(cmdArgs, " "))
	if args.TillerNamespace != "" {
		cmd = cmd.Env("TILLER_NAMESPACE", args.TillerNamespace)
	}
	var stderr strings.Builder
	if i.FailOnWarnings {
		cmd.CaptureStderr(&stderr)
	}
	return cmd.Run()
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
