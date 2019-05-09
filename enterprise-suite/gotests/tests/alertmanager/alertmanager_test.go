package alertmanager

import (
	"errors"
	"fmt"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/lbc"
	"strings"
	"testing"
	"time"

	"k8s.io/client-go/kubernetes/typed/apps/v1"

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
	alertm  *alertmanager.Connection

	depsClient      v1.DeploymentInterface
	oldConfigmap    *apiv1.ConfigMapVolumeSource
	alertmanagerDep *appsv1.Deployment

	configYaml = "../../resources/alertmanager.yml"
	configName = "alertmanager-smoke-config"
	appYaml    = "../../resources/alert-app.yaml"
	appName    = "es-alert-test"
)

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

	Context("warmup period", func() {
		const name = "my-warmingup-monitor"

		BeforeEach(func() {
			Expect(lbc.Install([]string{}, []string{"--set defaultMonitorWarmup=10m"})).To(Succeed(), "install with PVs")
			// delete in case it is lingering
			console.DeleteMonitor("es-alert-test/" + name)
			Expect(console.MakeAlertingMonitor("es-alert-test/" + name)).To(Succeed(), "should have created monitor")
		})

		AfterEach(func() {
			Expect(console.DeleteMonitor("es-alert-test/" + name)).To(Succeed())
		})

		It("should not generate data during the warmup period", func() {
			name := "my-warmingup-monitor"
			time.Sleep(5*time.Second)
			err := prom.HasNoData(fmt.Sprintf("model{name=\"%v\"}", name))
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("alerting", func() {
		setupAlert := func(name string) {
			err := console.MakeAlertingMonitor("es-alert-test/" + name)
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

		It("is firing in alertmanager", func() {
			name := "alert_monitor_alert"
			setupAlert(name)

			// Look for alert from our monitor using alertmanager api
			err := util.WaitUntilSuccess(util.LongWait, func() error {
				alerts, err := alertm.Alerts()
				Expect(err).ToNot(HaveOccurred())
				var alertNames []string
				var found bool
				for _, alert := range alerts {
					if val, ok := alert.Labels["alertname"]; ok {
						alertNames = append(alertNames, val)
						if val == name {
							found = true
							break
						}
					}
				}
				if !found {
					return fmt.Errorf("could not find alert %s in %s", name, alertNames)
				}

				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			//deleteAlert(name)
		})

		It("is firing in prometheus", func() {
			name := "alert_monitor_prom"
			setupAlert(name)

			// Look for alert from our monitor using prometheus query
			err := util.WaitUntilSuccess(util.LongWait, func() error {
				return prom.HasData(fmt.Sprintf(`ALERTS{alertname="%v",alertstate="firing",severity="warning"}`, name))
			})
			Expect(err).ToNot(HaveOccurred())

			//deleteAlert(name)
		})

		It("has correct generator URL", func() {
			name := "alert_monitor_generator_url"
			setupAlert(name)

			err := util.WaitUntilSuccess(util.LongWait, func() error {
				alerts, err := alertm.Alerts()
				if err != nil {
					return err
				}
				found := false
				for _, alert := range alerts {
					fmt.Fprintf(GinkgoWriter, "%v\n", alert)
					if strings.HasPrefix(alert.GeneratorURL, "http://console.test.bogus:30080") {
						found = true
						break
					}
				}
				if !found {
					return errors.New("expected alerts to have generator URL, but did not")
				}
				return nil
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

		It("is set correctly on Alertmanager", func() {
			pods, err := testenv.K8sClient.CoreV1().Pods(args.ConsoleNamespace).
				List(metav1.ListOptions{LabelSelector: "app.kubernetes.io/component=alertmanager"})
			Expect(err).ToNot(HaveOccurred())
			Expect(pods.Items).To(HaveLen(1))
			pod := pods.Items[0]

			var amContainer *apiv1.Container
			for _, c := range pod.Spec.Containers {
				if c.Name == "prometheus-alertmanager" {
					amContainer = &c
					break
				}
			}
			Expect(amContainer).ToNot(BeNil())
			Expect(amContainer.Args).To(ContainElement("--web.external-url=http://console.test.bogus:30080/service/alertmanager"))
		})
	})
})
