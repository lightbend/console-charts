package console

import (
	"fmt"

	"github.com/lightbend/console_test/util"
	"github.com/lightbend/console_test/util/kube"
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
		"./resources/app_with_service.yaml":                "es-test-via-service",
		"./resources/app_service_with_only_endpoints.yaml": "",
		"./resources/app_with_multiple_ports.yaml":         "es-test-with-multiple-ports",
	}

	var prom *prometheus.Connection
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
			Expect(prom.HasData(metric)).To(Equal(true))
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
			Expect(prom.HasData(fmt.Sprintf("model{name=\"%v\"}", metric))).To(Equal(true))
			Expect(prom.HasData(fmt.Sprintf("health{name=\"%v\"}", metric))).To(Equal(true))
		}

		// Coherency

		// Data with "es_workload" should also have a "namespace" label
		Expect(prom.HasData("count({es_workload=~\".+\", namespace=\"\", name!~\"node.*|kube_node.*\", __name__!~\"node.*|kube_node.*\"})")).To(Equal(false))
		// Health should have "es_workload" label, with a few known exceptions
		Expect(prom.HasData("health{es_workload=\"\", name!~\"node.*|kube_node.*|prometheus_target_down|scrape_time\"}")).To(Equal(false))
		// kube_pod_info must have es_workload labels
		Expect(prom.HasData("kube_pod_info{es_workload=~\".+\"}")).To(Equal(true))
		Expect(prom.HasData("kube_pod_info{es_workload=\"\"}")).To(Equal(false))
		// kube data mapped pod to workload labels
		Expect(prom.HasData("{__name__=~ \"kube_.+\", pod!=\"\", es_workload=\"\"}")).To(Equal(false))
		// All container data have a workload label
		Expect(prom.HasData("{__name__=~\"container_.+\", es_workload=\"\"}")).To(Equal(false))
		// All targets should be reachable
		Expect(prom.HasData("up{kubernetes_name != \"es-test-service-with-only-endpoints\"} == 1")).To(Equal(true))
		Expect(prom.HasData("up{kubernetes_name != \"es-test-service-with-only-endpoints\"} == 0")).To(Equal(false))

		// TODO:
		// kube-state-metrics
		// node-exporter
		// app tests
	})
})
