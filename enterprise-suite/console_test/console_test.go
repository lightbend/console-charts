package console_test

import (
	"testing"

	"github.com/lightbend/console_test/util/helm"
	"github.com/lightbend/console_test/util/minikube"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const consoleNamespace = "lightbend-test"

var _ = BeforeSuite(func() {
	Expect(minikube.IsRunning()).ShouldNot(BeTrue())
	Expect(minikube.Start(3, 6000)).To(Succeed())
	Expect(helm.Install(consoleNamespace)).To(Succeed())
})

var _ = AfterSuite(func() {
	Expect(minikube.IsRunning()).Should(BeTrue())
	Expect(minikube.Delete()).To(Succeed())
})

var _ = Describe("Console", func() {
	Context("Temporary", func() {
		It("works", func() {
			Expect(2 * 4).To(Equal(8))
		})
	})

	Context("Minikube", func() {
		It("is running", func() {
			Expect(minikube.IsRunning()).Should(BeTrue())
		})
	})

	Context("Helm", func() {
		It("is installed", func() {
			Expect(helm.IsInstalled()).Should(BeTrue())
		})
	})
})

func TestConsole(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Console Suite")
}
