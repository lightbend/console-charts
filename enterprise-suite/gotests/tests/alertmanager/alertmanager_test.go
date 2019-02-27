package alertmanager 

import (
	"fmt"
	"testing"

	"github.com/lightbend/gotests/args"
	"github.com/lightbend/gotests/testenv"

	"github.com/lightbend/gotests/util"
	"github.com/lightbend/gotests/util/kube"
	"github.com/lightbend/gotests/util/monitor"
	"github.com/lightbend/gotests/util/prometheus"
	"github.com/lightbend/gotests/util/alertmanager"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/typed/apps/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var prom *prometheus.Connection
var console *monitor.Connection
var alertm *alertmanager.Connection

var depsClient v1.DeploymentInterface
var oldConfigmap *apiv1.ConfigMapVolumeSource
var alertmanagerDep *appsv1.Deployment

var configYaml = "../../resources/alertmanager.yml"
var configName = "alertmanager-smoke-config"
var appYaml = "../../resources/alert-app.yaml"
var appName = "es-alert-test"

func TestAlertmanager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Alertmanager Suite")
}

var _ = BeforeSuite(func() {
	testenv.InitEnv()

	// Create test configmap
	err := kube.CreateConfigMap(args.ConsoleNamespace, configName, configYaml)
	Expect(err).ToNot(HaveOccurred())

	// Get deployments interface and alertmanager deployment
	depsClient = testenv.K8sClient.AppsV1().Deployments(args.ConsoleNamespace)
	alertmanagerDep, err = depsClient.Get("prometheus-alertmanager", metav1.GetOptions{})
	Expect(err).ToNot(HaveOccurred())

	// Change alertmanager configmap to our custom one, keep old one to revert back afterwards
	for _, volume := range alertmanagerDep.Spec.Template.Spec.Volumes {
		if volume.Name == "config-volume" {
			oldConfigmap = volume.ConfigMap
			volume.ConfigMap = &apiv1.ConfigMapVolumeSource{
				LocalObjectReference: apiv1.LocalObjectReference{
					Name: "alertmanager-smoke-config",
				},
			}
			break
		}
	}
	alertmanagerDep, err = depsClient.Update(alertmanagerDep)
	Expect(err).ToNot(HaveOccurred())

	// Deploy test app
	err = kube.ApplyYaml(args.ConsoleNamespace, appYaml)
	Expect(err).NotTo(HaveOccurred())

	// Wait for it to become ready
	err = util.WaitUntilSuccess(util.LongWait, func() error {
		return kube.IsDeploymentAvailable(testenv.K8sClient, args.ConsoleNamespace, appName)
	})
	Expect(err).NotTo(HaveOccurred())

	// Setup prometheus and console-api clients
	prom, err = prometheus.NewConnection(testenv.PrometheusAddr)
	Expect(err).To(Succeed())
	console, err = monitor.NewConnection(testenv.MonitorAPIAddr)
	Expect(err).To(Succeed())
	alertm, err = alertmanager.NewConnection(testenv.AlertmanagerAddr)

	// Wait for some scrapes to finish
	err = util.WaitUntilSuccess(util.LongWait, func() error {
		return prom.HasData("kube_pod_info")
	})
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	// Revert back to original alertmanager config
	for _, volume := range alertmanagerDep.Spec.Template.Spec.Volumes {
		if volume.Name == "config-volume" {
			volume.ConfigMap = oldConfigmap
		}
	}
	var err error
	alertmanagerDep, err = depsClient.Update(alertmanagerDep)
	Expect(err).ToNot(HaveOccurred())

	// Delete test configmap
	err = kube.DeleteConfigMap(args.ConsoleNamespace, configName)
	Expect(err).ToNot(HaveOccurred())

	// Delete test app
	err = kube.DeleteYaml(args.ConsoleNamespace, appYaml)
	Expect(err).ToNot(HaveOccurred())

	testenv.CloseEnv()
})

var _ = Describe("all:alertmanager", func() {
	Context("es-alert-test app integrity", func() {
		It("is running", func() {
			err := util.WaitUntilSuccess(util.SmallWait, func() error {
				return prom.HasData(fmt.Sprintf(`count( count by (instance) (ohai{es_workload="es-alert-test", namespace="%v"}) ) == 1`, args.ConsoleNamespace))
			})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("alerting", func() {
		It("can see alert firing", func() {
			// Setup alerting monitor
			err := console.MakeAlertingMonitor("es-alert-test/alerting_monitor", "up", 3)
			Expect(err).ToNot(HaveOccurred())
			err = util.WaitUntilSuccess(util.LongWait, func() error {
				return prom.HasModel("alerting_monitor")
			})
			Expect(err).ToNot(HaveOccurred())

			// Look for alert from our monitor
			alerts, err := alertm.Alerts()
			found := false
			for _, alert := range alerts {
				if val, ok := alert.Labels["alertname"]; ok && val == "alerting_monitor" {
					found = true
					break
				}
			}
			Expect(found).To(Equal(true))
		})
	})
})
