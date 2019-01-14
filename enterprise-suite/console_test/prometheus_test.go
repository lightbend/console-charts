package console

import (
	"fmt"

	"github.com/lightbend/console_test/util"
	"github.com/lightbend/console_test/util/kube"
	"github.com/lightbend/console_test/util/monitor"
	"github.com/lightbend/console_test/util/prometheus"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const promTestNamespace = "smoketest-prometheus"

var _ = Describe("all:prometheus", func() {
	// Resource yaml file to deployment name map
	appYamls := map[string]string{
		"./resources/app.yaml":                             "es-test",
		"./resources/app_with_service.yaml":                "es-test-via-service",
		"./resources/app_service_with_only_endpoints.yaml": "",
		"./resources/app_with_multiple_ports.yaml":         "es-test-with-multiple-ports",
	}

	var prom *prometheus.Connection
	var esMonitor *monitor.Connection
	var namespace *apiv1.Namespace

	// BeforeEach and AfterEach are used as single-time setup/teardown here because we have a
	// single It() and ginkgo supports only one global BeforeSuite/AfterSuite.

	BeforeEach(func() {
		var err error

		// Create namespace
		// TODO(mitkus): This is rather ugly code for such a simple operation, k8sClient could
		// benefit from a higher level wrapper with common operations.
		namespace = &apiv1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: promTestNamespace,
			},
		}
		_, err = k8sClient.CoreV1().Namespaces().Create(namespace)
		Expect(err).To(Succeed())

		// Create test app deployments + services
		for res, depName := range appYamls {
			err = kube.ApplyYaml(res, promTestNamespace)
			Expect(err).To(Succeed())

			// Wait for deployment to become ready
			if len(depName) > 0 {
				err = util.WaitUntilTrue(func() bool {
					// NOTE(mitkus): This might be far from best way to figure out if deployment is available
					dep, err := k8sClient.AppsV1().Deployments(promTestNamespace).Get(depName, metav1.GetOptions{})
					Expect(err).To(Succeed())
					if len(dep.Status.Conditions) > 0 {
						if dep.Status.Conditions[0].Type == "Available" {
							return true
						}
					}
					return false
				})
				Expect(err).To(Succeed())
			}
		}

		prom, err = prometheus.NewConnection(prometheusAddr)
		Expect(err).To(Succeed())

		esMonitor, err = monitor.NewConnection(monitorAPIAddr)
		Expect(err).To(Succeed())
	})

	AfterEach(func() {
		// Delete test app deployments + services
		for res := range appYamls {
			kube.DeleteYaml(res, promTestNamespace)

			// Ignore failures for now because deleting
			// 'es-test-service-with-only-endpoints' fails for some reason
			//Expect(err).To(Succeed())
		}

		// Delete now empty namespace
		err := k8sClient.CoreV1().Namespaces().Delete(promTestNamespace, &metav1.DeleteOptions{})
		Expect(err).To(Succeed())
	})

	It("works", func() {
		// Returns true if there were at least three scrapes of a given metric
		threeScrapes := func(metric string) bool {
			return prom.HasData(fmt.Sprintf("count_over_time(%v[10m]) > 2", metric))
		}

		// Wait until there's some scrapes finished
		util.WaitUntilTrue(func() bool {
			return threeScrapes("kube_pod_info")
		})
		util.WaitUntilTrue(func() bool {
			return threeScrapes("node_cpu")
		})

		// Prometheus

		promMetrics := []string{
			"prometheus_notifications_dropped_rate",
			"prometheus_notification_queue_percent",
			"prometheus_rule_evaluation_failures_rate",
			"prometheus_target_scrapes_exceeded_sample_limit_rate",
			"prometheus_tsdb_reloads_failures_rate",
			"prometheus_config_last_reload_successful",
			"prometheus_target_sync_percent",
		}

		for _, metric := range promMetrics {
			Expect(prom.HasData(metric)).To(BeTrue())
		}

		promHealth := []string{
			"prometheus_notifications_dropped",
			"prometheus_notification_queue",
			"prometheus_rule_evaluation_failures",
			"prometheus_target_too_many_metrics",
			"prometheus_tsdb_reloads_failures",
			"prometheus_target_down",
			"prometheus_config_reload_failed",
			"prometheus_scrape_time",
		}

		for _, metric := range promHealth {
			Expect(prom.HasData(fmt.Sprintf("model{name=\"%v\"}", metric))).To(BeTrue())
			Expect(prom.HasData(fmt.Sprintf("health{name=\"%v\"}", metric))).To(BeTrue())
		}

		// Coherency

		// Data with "es_workload" should also have a "namespace" label
		Expect(prom.HasData("count({es_workload=~\".+\", namespace=\"\", name!~\"node.*|kube_node.*\", __name__!~\"node.*|kube_node.*\"})")).To(BeFalse())
		// Health should have "es_workload" label, with a few known exceptions
		Expect(prom.HasData("health{es_workload=\"\", name!~\"node.*|kube_node.*|prometheus_target_down|scrape_time\"}")).To(BeFalse())
		// kube_pod_info must have es_workload labels
		Expect(prom.HasData("kube_pod_info{es_workload=~\".+\"}")).To(BeTrue())
		Expect(prom.HasData("kube_pod_info{es_workload=\"\"}")).To(BeFalse())
		// kube data mapped pod to workload labels
		Expect(prom.HasData("{__name__=~ \"kube_.+\", pod!=\"\", es_workload=\"\"}")).To(BeFalse())
		// All container data have a workload label
		Expect(prom.HasData("{__name__=~\"container_.+\", es_workload=\"\"}")).To(BeFalse())
		// All targets should be reachable
		Expect(prom.HasData("up{kubernetes_name != \"es-test-service-with-only-endpoints\"} == 1")).To(BeTrue())
		Expect(prom.HasData("up{kubernetes_name != \"es-test-service-with-only-endpoints\"} == 0")).To(BeFalse())

		// kube-state-metrics

		kubeStateMetrics := []string{
			"kube_pod_info",
			"kube_pod_ready",
			"kube_pod_container_restarts_rate",
			"kube_pod_failed",
			"kube_pod_not_running",
			"kube_workload_generation_lag",
		}

		for _, metric := range kubeStateMetrics {
			Expect(prom.HasData(metric)).To(BeTrue())
		}

		kubeStateHealth := []string{
			"kube_container_restarts",
			"kube_pod_not_ready",
			"kube_pod_not_running",
			"kube_workload_generation_lag",
		}

		for _, metric := range kubeStateHealth {
			Expect(prom.HasData(fmt.Sprintf("model{name=\"%v\"}", metric))).To(BeTrue())
			Expect(prom.HasData(fmt.Sprintf("health{name=\"%v\"}", metric))).To(BeTrue())
		}

		// TODO: Check kube-state-metrics logs

		// node-exporter

		nodeMetrics := []string{
			"node_cpu",
			"node_memory_MemAvailable",
			"node_filesystem_free",
			"node_network_transmit_errs",
		}

		for _, metric := range nodeMetrics {
			Expect(prom.HasData(metric)).To(BeTrue())
		}

		// TODO: Check node-exporter logs

		// App tests

		// App via 'Pod' service discovery
		appInstancesQuery := fmt.Sprintf("count( count by (instance) (ohai{es_workload=\"es-test\", namespace=\"%v\"}) ) == 2", promTestNamespace)
		err := util.WaitUntilTrue(func() bool {
			return prom.HasData(appInstancesQuery)
		})
		Expect(err).To(Succeed())

		// Pod discovery - automatic metrics should gain an es_monitor_type label when a custom monitor is created
		Expect(esMonitor.Make("es-test/my_custom_monitor", "up")).To(Succeed())
		err = util.WaitUntilTrue(func() bool {
			return prom.HasModel("my_custom_monitor")
		})
		Expect(err).To(Succeed())

		// App via 'Pod' service discovery with multiple metric ports
		appInstancesWithMultiplePortsQuery := fmt.Sprintf("count( count by (instance) (ohai{es_workload=\"es-test-with-multiple-ports\", namespace=\"%v\"}) ) == 4", promTestNamespace)
		err = util.WaitUntilTrue(func() bool {
			return prom.HasData(appInstancesWithMultiplePortsQuery)
		})
		Expect(err).To(Succeed())

		// App via 'Service' service discovery
		appInstancesViaServiceQuery := fmt.Sprintf("count( count by (instance) (ohai{es_workload=\"es-test-via-service\", namespace=\"%v\"}) ) == 2", promTestNamespace)
		err = util.WaitUntilTrue(func() bool {
			return prom.HasData(appInstancesViaServiceQuery)
		})
		Expect(err).To(Succeed())

		// 'Service' discovery - automatic metrics should gain an es_monitor_type label when a custom monitor is created
		Expect(esMonitor.Make("es-test-via-service/my_custom_monitor_for_service", "up")).To(Succeed())
		err = util.WaitUntilTrue(func() bool {
			return prom.HasModel("my_custom_monitor_for_service")
		})
		Expect(err).To(Succeed())
		Expect(prom.HasData("up{es_workload=\"es-test-via-service\", es_monitor_type=\"es-test-via-service\"}")).To(BeTrue())

		// 'Service' discovery with only endpoints
		err = util.WaitUntilTrue(func() bool {
			return prom.HasData(fmt.Sprintf("count( count by (instance) (up{ "+
				"job=\"kubernetes-services\", kubernetes_name=\"es-test-service-with-only-endpoints\", namespace=\"%v\""+
				"}) ) == 1", promTestNamespace))
		})
		Expect(err).To(Succeed())

		// kubernetes-cadvisor metrics should have an es_monitor_type label.
		Expect(esMonitor.Make("es-test/es-monitor-type-test", "container_cpu_load_average_10s")).To(Succeed())
		err = util.WaitUntilTrue(func() bool {
			return prom.HasModel("es-monitor-type-test")
		})
		Expect(err).To(Succeed())

		// Specific test for regression of es-backend/issues/430.
		Expect(prom.HasData("{job=\"kubernetes-cadvisor\",es_monitor_type=\"es-test\"}")).To(BeTrue())

		// Succeeds if all data for the workload es-test  has a matching es_monitor_type
		// Note we're currently ignoring health metrics because 'bad' data can stick around for 15m given their time window.
		err = util.WaitUntilTrue(func() bool {
			return prom.HasData("{es_workload=\"es-test\", es_monitor_type!=\"es-test\", __name__!=\"health\"}")
		})
		Expect(err).To(Succeed())
	})
})
