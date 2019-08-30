package console

import (
	"strings"
	"testing"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/args"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/testenv"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/kube"

	"regexp"

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
		logs, err := kube.GetLogs(args.ConsoleNamespace, "app.kubernetes.io/component=grafana")
		Expect(err).To(Succeed(), "should have retrieved logs")

		logLines := strings.Split(logs, "\n")
		// Error messages we can safely ignore:
		// (1) Alert Rule Result Error - happens when Prometheus isn't initialized yet - the Kafka alert in the Strimzi
		// dashboard fails.
		whitelist := regexp.MustCompile(`Alert Rule Result Error`)

		var filteredLogLines []string
		for _, line := range logLines {
			if strings.Contains(line, "lvl=eror") {
				// this is an error line
				if !whitelist.MatchString(line) {
					filteredLogLines = append(filteredLogLines, line)
				}
			}
		}

		Expect(filteredLogLines).To(BeEmpty(), "should not have any error lines")
	})
})
