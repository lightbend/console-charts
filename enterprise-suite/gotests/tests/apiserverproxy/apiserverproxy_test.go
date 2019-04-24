package apiserverproxy

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/args"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/testenv"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/urls"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var (
	proxyPort int
	proxyCmd  *util.CmdBuilder
)

var _ = BeforeSuite(func() {
	proxyPort = util.FindFreePort()
	proxyCmd = util.Cmd("kubectl", "proxy", "-p", strconv.Itoa(proxyPort))
	testenv.InitEnv()
	Expect(proxyCmd.StartAsync()).To(Succeed())
})

var _ = AfterSuite(func() {
	Expect(proxyCmd.StopAsync()).To(Succeed())
	testenv.CloseEnv()
})

var _ = Describe("all:apiserverproxy", func() {
	DescribeTable("can access via apiserver proxy", func(serviceName, servicePath string) {
		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/namespaces/%s/services/%s:http/proxy%s",
			proxyPort, args.ConsoleNamespace, serviceName, servicePath)
		By(url)
		_, err := urls.Get200(url)
		Expect(err).ToNot(HaveOccurred())
	},
		Entry("console-server", "console-server", "/"),
		Entry("console-api", "console-api", "/status"),
		Entry("prometheus-server", "prometheus-server", "/-/healthy"),
		Entry("grafana-server", "grafana-server", "/api/org"),
		Entry("alertmanager", "alertmanager", "/-/healthy"),
	)
})

func TestApiserverProxy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Apiserver Proxy Suite")
}
