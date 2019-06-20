package ingress

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/args"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/testenv"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	consoleRemotePort = 8080
	consoleDeployment = "deployment/console-frontend"
)

var (
	portForwardCmd *util.CmdBuilder
	localPort      int
)

func TestPortForward(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Port Forward Suite")
}

var _ = BeforeSuite(func() {
	testenv.InitEnv()

	localPort = util.FindFreePort()
	portForwardCmd = util.Cmd("kubectl", "port-forward",
		"-n", args.ConsoleNamespace,
		consoleDeployment,
		fmt.Sprintf("%d:%d", localPort, consoleRemotePort))
	Expect(portForwardCmd.StartAsync()).To(Succeed())
})

var _ = AfterSuite(func() {
	Expect(portForwardCmd.StopAsync()).To(Succeed())
	testenv.CloseEnv()
})

var _ = Describe("all:portforward", func() {
	It("forwards 127.0.0.1 requests to console", func() {
		addr := fmt.Sprintf("http://127.0.0.1:%v", localPort)

		err := util.WaitUntilSuccess(util.LongWait, func() error {
			client := &http.Client{
				Timeout: 10 * time.Second,
			}
			resp, err := client.Get(addr)
			if err != nil {
				return err
			}

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			if resp.StatusCode != 200 {
				return fmt.Errorf("wanted 200 response, got %d: %s", resp.StatusCode, string(body))
			}

			return nil
		})
		Expect(err).ToNot(HaveOccurred())
	})
})
