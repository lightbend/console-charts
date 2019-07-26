package alertmanager

import (
	"fmt"
	"testing"

	gov1 "k8s.io/client-go/kubernetes/typed/apps/v1"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/args"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/testenv"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/util"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/kube"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/monitor"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/prometheus"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAlertmanager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Alertmanager Suite")
}

var (
	prom    *prometheus.Connection
	console *monitor.Connection

	depsClient      gov1.DeploymentInterface

	appYaml    = "../../resources/alert-app.yaml"
	appName    = "es-alert-test"
)

var _ = BeforeSuite(func() {
	testenv.InitEnv()

	// Get deployments interface and alertmanager deployment
	depsClient = testenv.K8sClient.AppsV1().Deployments(args.ConsoleNamespace)

	// Deploy test app
	err := kube.ApplyYaml(args.ConsoleNamespace, appYaml)
	Expect(err).NotTo(HaveOccurred())

	// Wait for it to become ready
	err = util.WaitUntilSuccess(util.LongWait, func() error {
		return kube.IsDeploymentAvailable(testenv.K8sClient, args.ConsoleNamespace, appName)
	})
	Expect(err).NotTo(HaveOccurred())

	// Setup prometheus and console-api clients
	prom, err = prometheus.NewConnection(testenv.PrometheusAddr)
	Expect(err).To(Succeed())
	console, err = monitor.NewConnection(testenv.ConsoleAPIAddr)
	Expect(err).To(Succeed())

	// Wait for some scrapes to finish
	err = util.WaitUntilSuccess(util.LongWait, func() error {
		return prom.HasData("kube_pod_info")
	})
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	if depsClient != nil {
		// Delete test app
		err := kube.DeleteYaml(args.ConsoleNamespace, appYaml)
		Expect(err).ToNot(HaveOccurred())
	}

	testenv.CloseEnv()
})

var _ = Describe("all:alertmanager", func() {
	Context("es-alert-test app", func() {
		It("has visible metrics", func() {
			err := util.WaitUntilSuccess(util.LongWait, func() error {
				return prom.HasData(`count( count by (instance) (ohai{es_workload="es-alert-test", namespace="%v"}) ) == 1`, args.ConsoleNamespace)
			})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("warmup period", func() {
		const name = "my-warmingup-monitor"

		BeforeEach(func() {
			// delete in case it is lingering
			console.TryDeleteMonitor("es-alert-test/" + name)
			err := util.WaitUntilSuccess(util.LongWait, func() error {
				return prom.HasNoData(`model{name="%s"}`, name)
			})
			Expect(err).ToNot(HaveOccurred(), "should have successfully removed lingering monitor %s", name)
			Expect(console.MakeAlertingMonitor("es-alert-test/"+name, "10m")).To(Succeed(), "should have created monitor")
		})

		AfterEach(func() {
			Expect(console.DeleteMonitor("es-alert-test/" + name)).To(Succeed())
		})

		It("should not generate data during the warmup period", func() {
			name := "my-warmingup-monitor"

			err := util.WaitUntilSuccess(util.LongWait, func() error {
				return prom.HasData(`model{name="%s"}`, name)
			})
			Expect(err).ToNot(HaveOccurred(), "should have gotten model data for %s", name)

			err = prom.HasNoData(`health{name="%s"}`, name)
			Expect(err).ToNot(HaveOccurred(), "should not have gotten health data for %s", name)
		})
	})

	Context("alerting", func() {
		setupAlert := func(name string) {
			err := console.MakeAlertingMonitor("es-alert-test/"+name, "1s")
			Expect(err).ToNot(HaveOccurred(), "should have created monitor")
			err = util.WaitUntilSuccess(util.LongWait, func() error {
				return prom.HasModel(name)
			})
			Expect(err).ToNot(HaveOccurred())
		}

		deleteAlert := func(name string) {
			err := console.DeleteMonitor("es-alert-test/" + name)
			Expect(err).ToNot(HaveOccurred(), "should have deleted monitor")
		}

		It("is firing in prometheus", func() {
			name := "alert_monitor_prom"
			setupAlert(name)

			// Look for alert from our monitor using prometheus query
			err := util.WaitUntilSuccess(util.LongWait, func() error {
				return prom.HasData(fmt.Sprintf(`ALERTS{alertname="%v",alertstate="firing",severity="warning"}`, name))
			})
			Expect(err).ToNot(HaveOccurred())

			deleteAlert(name)
		})
	})

	Context("external URL", func() {
		It("is set correctly on Prometheus", func() {
			pods, err := testenv.K8sClient.CoreV1().Pods(args.ConsoleNamespace).
				List(metav1.ListOptions{LabelSelector: "app.kubernetes.io/component=console-backend"})
			Expect(err).ToNot(HaveOccurred())
			Expect(pods.Items).To(HaveLen(1))
			pod := pods.Items[0]

			var promContainer *apiv1.Container
			for _, c := range pod.Spec.Containers {
				if c.Name == "prometheus-server" {
					promContainer = &c
					break
				}
			}
			Expect(promContainer).ToNot(BeNil())
			Expect(promContainer.Args).To(ContainElement("--web.external-url=http://console.test.bogus:30080/service/prometheus"))
		})
	})
})
