package apiserverproxy

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"

	"github.com/lightbend/gotests/args"
	"github.com/lightbend/gotests/util"

	"github.com/lightbend/gotests/testenv"

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

		err := util.WaitUntilSuccess(func() error {
			resp, err := http.Get(url)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			if resp.StatusCode != 200 {
				return fmt.Errorf("wanted 200, got %d: %s", resp.StatusCode, string(body))
			}

			return nil
		}, util.SmallWait)

		Expect(err).ToNot(HaveOccurred())
	},
		Entry("console-server", "console-server", "/"),
		Entry("es-monitor-api", "es-monitor-api", "/status"),
		Entry("prometheus-server", "prometheus-server", "/-/healthy"),
		Entry("grafana-server", "grafana-server", "/api/org"),
		Entry("alertmanager", "alertmanager", "/-/healthy"),
	)
})

func TestApiserverProxy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Apiserver Proxy Suite")
}
