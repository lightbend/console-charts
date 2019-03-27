package console

import (
	"testing"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/testenv"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/kube"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGrafana(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Console Suite")
}

var _ = BeforeSuite(func() {
	testenv.InitEnv()
})

var _ = AfterSuite(func() {
	testenv.CloseEnv()
})

var _ = Describe("all:grafana", func() {
	It("loads all of the dashboards on start up without error", func() {
		Expect(kube.GetLogs(testenv.ConsoleNamespace, "app.kubernetes.io/component=grafana")).
			ToNot(ContainSubstring("Failed to auto update app dashboard"))
	})
})
