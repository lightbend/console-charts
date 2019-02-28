package testenv

import (
	"fmt"
	"os"
	"time"

	"github.com/onsi/ginkgo"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/args"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/helm"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/kube"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/lbc"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/minikube"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/oc"

	. "github.com/onsi/gomega"
)

var (
	// The following variables are used across tests to access kubernetes
	// and make requests on Console components.

	K8sClient        *kubernetes.Clientset
	ConsoleAddr      string
	PrometheusAddr   string
	MonitorAPIAddr   string
	GrafanaAddr      string
	AlertmanagerAddr string

	isMinikube  bool
	isOpenshift bool

	openshiftConsoleService = "console-server"
	helmReleaseName         = "enterprise-suite"
	consolePVCs             = []string{"prometheus-storage", "es-grafana-storage", "alertmanager-storage"}

	foundStorageClass  = false
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

	if args.Cleanup && (helm.ReleaseExists(helmReleaseName)) {
		// Cleanup with allowFailures=true
		cleanup(true)
	}

	additionalArgs := []string{"--set esConsoleURL=http://console.test.bogus:30080"}
	if isMinikube {
		additionalArgs = append(additionalArgs, "--set exposeServices=NodePort")
		foundStorageClass = true
	} else {
		// Look for expected storage classes, run with emptyDir if they don't exist
		for _, storageClass := range []string{"standard", "gp2"} {
			if kube.StorageClassExists(K8sClient, storageClass) {
				foundStorageClass = true
				additionalArgs = append(additionalArgs,
					fmt.Sprintf("--set usePersistentVolumes=true,defaultStorageClass=%v", storageClass))
				break
			}
		}
		if !foundStorageClass {
			additionalArgs = append(additionalArgs, "--set usePersistentVolumes=false,managePersistentVolumes=false")
		}
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
	if err := lbc.Install(args.ConsoleNamespace, additionalArgs...); err != nil {
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

		ConsoleAddr = fmt.Sprintf("http://%v", addr)
	}

	PrometheusAddr = fmt.Sprintf("%v/service/prometheus", ConsoleAddr)
	MonitorAPIAddr = fmt.Sprintf("%v/service/console-api", ConsoleAddr)
	GrafanaAddr = fmt.Sprintf("%v/service/grafana", ConsoleAddr)
	AlertmanagerAddr = fmt.Sprintf("%v/service/alertmanager", ConsoleAddr)

	testEnvInitialized = true
}

func CloseEnv() {
	if !testEnvInitialized {
		return
	}

	if args.Cleanup {
		cleanup(false)
	}

	testEnvInitialized = false
}

func cleanup(allowFailures bool) {
	fmt.Fprintf(ginkgo.GinkgoWriter, "Cleaning up old installation...")
	if isOpenshift {
		if err := oc.Unexpose(openshiftConsoleService); err != nil && !allowFailures {
			Expect(err).To(Succeed(), "oc delete route")
		}
	}

	// Uninstall Console using helm
	if err := lbc.Uninstall(); err != nil && !allowFailures {
		Expect(err).To(Succeed(), "lbc.Uninstall")
	}

	fmt.Fprintln(ginkgo.GinkgoWriter, "Done cleaning up old installation")
}
