package console

import (
	"fmt"
	"testing"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/lightbend/console_test/args"

	"github.com/lightbend/console_test/util/helm"
	"github.com/lightbend/console_test/util/lbc"
	"github.com/lightbend/console_test/util/minikube"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// The following variables are used in this package to access kubernetes
// and make requests on Console components.
var k8sClient *kubernetes.Clientset
var consoleAddr string
var prometheusAddr string
var monitorApiAddr string

var _ = BeforeSuite(func() {
	// Start minikube if --start-minikube arg was given
	if args.StartMinikube {
		Expect(minikube.IsRunning()).ShouldNot(BeTrue())
		Expect(minikube.Start(3, 6000)).To(Succeed())
	}

	// Setup k8s client
	config, err := clientcmd.BuildConfigFromFlags("", args.Kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	k8sClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// If we started minikube, install helm as well
	if args.StartMinikube {
		Expect(helm.Install(k8sClient)).To(Succeed())
	}

	// Install console
	Expect(lbc.Install(args.ConsoleNamespace)).To(Succeed())

	// Setup addresses for making requests to Console components
	if minikube.IsRunning() {
		ip, err := minikube.Ip()
		Expect(err).To(Succeed())

		consoleAddr = fmt.Sprintf("http://%v:30080", ip)
		prometheusAddr = fmt.Sprintf("%v/service/prometheus", consoleAddr)
		monitorApiAddr = fmt.Sprintf("%v/service/es-monitor-api", consoleAddr)
	} else {
		// TODO: Setup addresses for openshift
	}
})

var _ = AfterSuite(func() {
	Expect(lbc.Uninstall()).To(Succeed())
	if args.StartMinikube {
		Expect(minikube.IsRunning()).Should(BeTrue())
		Expect(minikube.Delete()).To(Succeed())
	}
})

var _ = Describe("minikube:verify", func() {
	Context("Minikube", func() {
		It("is running", func() {
			Expect(minikube.IsRunning()).Should(BeTrue())
		})
	})
})

var _ = Describe("all:verify", func() { 
	Context("Helm", func() {
		It("is installed", func() {
			Expect(helm.IsInstalled()).Should(BeTrue())
		})
	})

	Context("Console", func() {
		It("is verified", func() {
			Expect(lbc.Verify(args.ConsoleNamespace)).Should(Succeed())
		})
	})
})

func TestConsole(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Console Suite")
}
