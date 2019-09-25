package console

import (
	"fmt"
	"testing"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/args"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/testenv"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/kube"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/lbc"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/urls"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	// ES can take a while to start up
	installer.HelmWait = "240"
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

var _ = Describe("minikube:elasticsearch", func() {
	It("has data volumes owned by Elasticsearch", func() {
		pods, err := testenv.K8sClient.CoreV1().Pods(args.ConsoleNamespace).
			List(metav1.ListOptions{LabelSelector: "app.kubernetes.io/component=console-backend"})
		Expect(err).To(BeNil())
		pod := pods.Items[0]
		Expect(kube.FileOwner(args.ConsoleNamespace, pod.Name, "elasticsearch", "/usr/share/elasticsearch/data")).To(Equal("1000"))
	})
})
