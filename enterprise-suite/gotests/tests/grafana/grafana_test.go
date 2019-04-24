package console

import (
	"testing"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/args"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/testenv"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/kube"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGrafana(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Grafana Suite")
}

var _ = BeforeSuite(func() {
	testenv.InitEnv()
})

var _ = AfterSuite(func() {
	testenv.CloseEnv()
})

var _ = Describe("all:grafana", func() {
	It("loads all of the dashboards on start up without error", func() {
		// Exact message is:
		// t=2019-03-26T12:48:10+0000 lvl=eror msg="Failed to auto update app dashboard" logger=plugins pluginId=cinnamon-prometheus-app error="Dashboard not found"
		// We test for "lvl=eror" in case Grafana changes this error message.
		Expect(kube.GetLogs(args.ConsoleNamespace, "app.kubernetes.io/component=grafana")).
			ToNot(ContainSubstring("lvl=eror"))
	})
})
