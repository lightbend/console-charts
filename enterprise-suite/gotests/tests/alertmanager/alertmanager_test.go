package alertmanager

import (
	"fmt"
	"strings"
	"testing"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/args"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/testenv"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/util"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/alertmanager"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/kube"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/monitor"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/prometheus"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"

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

	// Delete test configmap if it existed previously, ignore failure
	kube.DeleteConfigMap(args.ConsoleNamespace, configName)

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
	console, err = monitor.NewConnection(testenv.ConsoleAPIAddr)
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
	if alertmanagerDep != nil {
		for _, volume := range alertmanagerDep.Spec.Template.Spec.Volumes {
			if volume.Name == "config-volume" {
				volume.ConfigMap = oldConfigmap
			}
		}
	}

	if depsClient != nil {
		_, err := depsClient.Update(alertmanagerDep)
		Expect(err).ToNot(HaveOccurred())

		// Delete test configmap
		err = kube.DeleteConfigMap(args.ConsoleNamespace, configName)
		Expect(err).ToNot(HaveOccurred())

		// Delete test app
		err = kube.DeleteYaml(args.ConsoleNamespace, appYaml)
		Expect(err).ToNot(HaveOccurred())
	}

	testenv.CloseEnv()
})

var _ = Describe("all:alertmanager", func() {
	Context("es-alert-test app", func() {
		It("has visible metrics", func() {
			err := util.WaitUntilSuccess(util.LongWait, func() error {
				return prom.HasData(fmt.Sprintf(`count( count by (instance) (ohai{es_workload="es-alert-test", namespace="%v"}) ) == 1`, args.ConsoleNamespace))
			})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("alerting", func() {
		setupAlert := func(name string) {
			err := console.MakeAlertingMonitor("es-alert-test/"+name, "up", 3)
			Expect(err).ToNot(HaveOccurred())
			err = util.WaitUntilSuccess(util.LongWait, func() error {
				return prom.HasModel(name)
			})
			Expect(err).ToNot(HaveOccurred())
		}

		deleteAlert := func(name string) {
			err := console.DeleteMonitor("es-alert-test/" + name)
			Expect(err).ToNot(HaveOccurred())
		}

		It("is firing in alertmanager", func() {
			name := "alert_monitor_alertm"
			setupAlert(name)

			// Look for alert from our monitor using alertmanager api
			alerts, err := alertm.Alerts()
			Expect(err).ToNot(HaveOccurred())
			var alertNames []string
			for _, alert := range alerts {
				if val, ok := alert.Labels["alertname"]; ok {
					alertNames = append(alertNames, val)
				}
			}
			Expect(alertNames).To(ContainElement(name))

			deleteAlert(name)
		})

		It("is firing in prometheus", func() {
			name := "alert_monitor_prom"
			setupAlert(name)

			// Look for alert from our monitor using prometheus query
			err := prom.HasData(fmt.Sprintf(`ALERTS{alertname="%v",alertstate="firing",severity="warning"}`, name))
			Expect(err).ToNot(HaveOccurred())

			deleteAlert(name)
		})

		It("has correct generator URL", func() {
			name := "alert_monitor_generator_url"
			setupAlert(name)

			alerts, err := alertm.Alerts()
			Expect(err).ToNot(HaveOccurred())
			found := false
			for _, alert := range alerts {
				if strings.HasPrefix(alert.GeneratorURL, "http://console.test.bogus:30080") {
					found = true
					break
				}
			}
			Expect(found).To(Equal(true))

			deleteAlert(name)
		})
	})
})
