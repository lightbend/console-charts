package testenv

import (
	"fmt"
	"os"
	"time"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/args"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/util"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/lbc"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/minikube"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/oc"

	. "github.com/onsi/gomega"
)

var (
	// The following variables are used across tests to access kubernetes
	// and make requests on Console components.

	K8sClient            *kubernetes.Clientset
	ConsoleAddr          string
	PrometheusAddr       string
	ConsoleAPIAddr       string
	GrafanaAddr          string
	LegacyMonitorAPIAddr string

	isMinikube  bool
	isOpenshift bool

	openshiftConsoleService = "console-server"
	helmReleaseName         = "enterprise-suite"
	consolePVCs             = []string{"prometheus-storage", "es-grafana-storage"}

	testEnvInitialized = false
)

func InitEnv() {
	if testEnvInitialized {
		return
	}

	// This detection is not 100% accurate, but should be enough for testing on two different platforms
	isMinikube = minikube.IsRunning()
	if !isMinikube {
		isOpenshift = oc.IsRunning()
	}
	if !isMinikube && !isOpenshift {
		panic("Could not determine which kubernetes platform is being used")
	}

	// Setup k8s client
	config, err := clientcmd.BuildConfigFromFlags("", args.Kubeconfig)
	if err != nil {
		Expect(err).To(Succeed(), "new k8sclient config")
	}
	K8sClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		Expect(err).To(Succeed(), "new k8sclient")
	}

	// Create an async goroutine to report when install is taking a while.
	ticker := time.NewTicker(15 * time.Second)
	start := time.Now()
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				if _, err := fmt.Fprintf(os.Stdout, "\nlbc.py installation is still ongoing [%.0fs]...", time.Now().Sub(start).Seconds()); err != nil {
					panic(err)
				}
			case <-done:
				if _, err := fmt.Fprintf(os.Stdout, "\n"); err != nil {
					panic(err)
				}
				return
			}
		}
	}()
	defer close(done)
	defer ticker.Stop()

	// Install console
	if err := lbc.DefaultInstaller().Install(); err != nil {
		Expect(err).To(Succeed(), "lbc.Install")
	}

	// Setup console address for making requests to Console components
	if isMinikube {
		ip, err := minikube.Ip()
		if err != nil {
			Expect(err).To(Succeed(), "unable to get minikube ip")
		}

		ConsoleAddr = fmt.Sprintf("http://%v:30080", ip)

	}
	if isOpenshift {
		err := oc.Expose(openshiftConsoleService)
		if err != nil {
			panic(fmt.Sprintf("unable to expose openshift service: %v", err))
		}

		addr, err := oc.Address(openshiftConsoleService)
		if err != nil {
			panic(fmt.Sprintf("unable to get openshift address: %v", err))
		}

		ConsoleAddr = fmt.Sprintf("https://%v", addr)
	}

	PrometheusAddr = fmt.Sprintf("%v/service/prometheus", ConsoleAddr)
	ConsoleAPIAddr = fmt.Sprintf("%v/service/console-api", ConsoleAddr)
	GrafanaAddr = fmt.Sprintf("%v/service/grafana", ConsoleAddr)
	LegacyMonitorAPIAddr = fmt.Sprintf("%v/service/es-monitor-api", ConsoleAddr)

	testEnvInitialized = true

	util.LogDebugInfo()
}

func CloseEnv() {
	if !testEnvInitialized {
		return
	}

	testEnvInitialized = false
}
