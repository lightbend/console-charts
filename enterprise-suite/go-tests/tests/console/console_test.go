package console

import (
	"testing"

	"github.com/lightbend/go_tests/args"
	"github.com/lightbend/go_tests/testenv"

	"github.com/lightbend/go_tests/util/helm"
	"github.com/lightbend/go_tests/util/lbc"
	"github.com/lightbend/go_tests/util/minikube"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = BeforeSuite(func() {
	testenv.InitEnv()	
})

var _ = AfterSuite(func() {
	testenv.CloseEnv()	
})

// The most basic verification tests

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
