package console

import (
	"fmt"
	"testing"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/testenv"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/lbc"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/urls"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGrafana(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Elasticsearch Suite")
}

var _ = BeforeSuite(func() {
	testenv.InitEnv()
	installer := lbc.DefaultInstaller()
	installer.AdditionalHelmArgs = []string{"--set enableElasticsearch=true"}
	Expect(installer.Install()).To(Succeed(), "install with Elasticsearch enabled")
})

var _ = AfterSuite(func() {
	testenv.CloseEnv()
})

var _ = Describe("all:elasticsearch", func() {
	It("can access Elasticsearch", func() {
		addr := fmt.Sprintf("%v/%v", testenv.ConsoleAddr, "service/elasticsearch/")
		_, err := urls.Get200(addr)
		Expect(err).To(BeNil(), addr)
	})
})
