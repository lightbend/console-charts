package testenv

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/lightbend/gotests/args"

	"github.com/lightbend/gotests/util/lbc"
	"github.com/lightbend/gotests/util/minikube"
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

	// Setup k8s client
	config, err := clientcmd.BuildConfigFromFlags("", args.Kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	K8sClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
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

	testEnvInitialized = false
}
