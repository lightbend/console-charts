package console

import (
	"testing"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/args"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/monitor"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/testenv"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/helm"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/lbc"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/minikube"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/oc"

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

var _ = Describe("openshift:verify", func() {
	Context("Openshift", func() {
		It("is running", func() {
			Expect(oc.IsRunning()).Should(BeTrue())
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
		It("is verified to be installed correctly", func() {
			Expect(lbc.Verify(args.ConsoleNamespace, args.TillerNamespace)).Should(Succeed())
		})

		It("can access the legacy es-monitor-api endpoint still", func() {
			_, err := monitor.NewConnection(testenv.LegacyMonitorAPIAddr)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

func TestConsole(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Console Suite")
}
