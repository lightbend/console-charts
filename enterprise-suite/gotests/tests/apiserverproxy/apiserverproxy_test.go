package apiserverproxy

import (
	"fmt"
	"testing"

	"github.com/lightbend/gotests/util"

	"github.com/lightbend/gotests/testenv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	proxyCmd *util.CmdBuilder
)

var _ = BeforeSuite(func() {
	proxyCmd = util.Cmd("kubectl", "proxy")
	testenv.InitEnv()
	Expect(proxyCmd.StartAsync()).To(Succeed())
})

var _ = AfterSuite(func() {
	Expect(proxyCmd.StopAsync()).ShouldNot(HaveOccurred())
	testenv.CloseEnv()
})

var _ = Describe("all:apiserverproxy", func() {
	It("can access Console", func() {

	})
})

func TestIngress(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ingress Suite")
}

// Returns base URL to access service $1 at port $2 via the proxy.
func getProxyURL(kubectlProxyAddr, namespace, service, servicePort string) string {
	return fmt.Sprintf("http://%s/api/v1/namespaces/%s/services/%s:%s/proxy",
		kubectlProxyAddr, namespace, service, servicePort)
}
