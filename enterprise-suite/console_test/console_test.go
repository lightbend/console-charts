package console_test

import (
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

const consoleNamespace = "lightbend-test"

var k8sClient *kubernetes.Clientset

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
	Expect(lbc.Install(consoleNamespace)).To(Succeed())
})

var _ = AfterSuite(func() {
	Expect(lbc.Uninstall()).To(Succeed())
	if args.StartMinikube {
		Expect(minikube.IsRunning()).Should(BeTrue())
		Expect(minikube.Delete()).To(Succeed())
	}
})

var _ = Describe("Console", func() {
	Context("Temporary", func() {
		It("works", func() {
			Expect(2 * 4).To(Equal(8))
		})
	})

	Context("Minikube", func() {
		It("is running", func() {
			if !args.Minikube {
				return
			}
			Expect(minikube.IsRunning()).Should(BeTrue())
		})
	})

	Context("Helm", func() {
		It("is installed", func() {
			Expect(helm.IsInstalled()).Should(BeTrue())
		})
	})

	Context("Console", func() {
		It("is verified", func() {
			Expect(lbc.Verify(consoleNamespace)).Should(Succeed())
		})
	})
})

func TestConsole(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Console Suite")
}
