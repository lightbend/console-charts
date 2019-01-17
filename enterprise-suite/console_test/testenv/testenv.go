package testenv

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/lightbend/console_test/args"

	"github.com/lightbend/console_test/util/helm"
	"github.com/lightbend/console_test/util/lbc"
	"github.com/lightbend/console_test/util/minikube"
)

var (
	// The following variables are used accross tests to access kubernetes
	// and make requests on Console components.
	K8sClient *kubernetes.Clientset
	ConsoleAddr string
	PrometheusAddr string
	MonitorAPIAddr string
	GrafanaAddr string

	testEnvInitialized = false
)

func InitEnv() {
	if testEnvInitialized {
		return
	}

	// Start minikube if --start-minikube arg was given
	if args.StartMinikube {
		if minikube.IsRunning() {
			panic("minikube appears to be already running, try running without --start-minikube flag")
		}
		if err := minikube.Start(3, 6000); err != nil {
			panic(err.Error())
		}
	}

	// Setup k8s client
	config, err := clientcmd.BuildConfigFromFlags("", args.Kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	K8sClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// If we started minikube, install helm as well
	if args.StartMinikube {
		if err := helm.Install(K8sClient); err != nil {
			panic(err.Error())
		}
	}

	// Install console
	if err := lbc.Install(args.ConsoleNamespace); err != nil {
		panic(err.Error())
	}

	// Setup addresses for making requests to Console components
	if minikube.IsRunning() {
		ip, err := minikube.Ip()
		if err != nil {
			panic(fmt.Sprintf("unable to get minikube ip: %v", err))
		}

		ConsoleAddr = fmt.Sprintf("http://%v:30080", ip)
		PrometheusAddr = fmt.Sprintf("%v/service/prometheus", ConsoleAddr)
		MonitorAPIAddr = fmt.Sprintf("%v/service/es-monitor-api", ConsoleAddr)
		GrafanaAddr = fmt.Sprintf("%v/service/grafana", ConsoleAddr)
	} else {
		// TODO: Setup addresses for openshift
	}

	testEnvInitialized = true
}

func CloseEnv() {
	if !testEnvInitialized {
		return
	}

	if err := lbc.Uninstall(); err != nil {
		panic(err.Error())
	}

	if args.StartMinikube {
		if !minikube.IsRunning() {
			panic("expected minikube to be running")
		}
		if err := minikube.Delete(); err != nil {
			panic(err.Error())
		}
	}

	testEnvInitialized = false
}
