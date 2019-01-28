package prometheus

import (
	"fmt"
	"testing"

	"github.com/lightbend/gotests/args"
	"github.com/lightbend/gotests/testenv"
	"github.com/lightbend/gotests/util"
	"github.com/lightbend/gotests/util/kube"
	"github.com/lightbend/gotests/util/monitor"
	"github.com/lightbend/gotests/util/prometheus"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var prom *prometheus.Connection
var esMonitor *monitor.Connection

// Resource yaml file to deployment name map
var	appYamls = map[string]string{
	"../../resources/app.yaml":                             "es-test",
	"../../resources/app_with_service.yaml":                "es-test-via-service",
	"../../resources/app_service_with_only_endpoints.yaml": "",
	"../../resources/app_with_multiple_ports.yaml":         "es-test-with-multiple-ports",
}

var _ = BeforeSuite(func() {
		testenv.InitEnv()

		var err error

		// Create test app deployments + services
		for res, depName := range appYamls {
			err = kube.ApplyYaml(args.ConsoleNamespace, res)
			Expect(err).To(Succeed())

			// Wait for deployment to become ready
			if len(depName) > 0 {
				err = util.WaitUntilTrue(func() bool {
					return kube.IsDeploymentAvailable(testenv.K8sClient, args.ConsoleNamespace, depName)
				}, fmt.Sprintf("deployment %v did not become available", depName))
				Expect(err).To(Succeed())
			}
		}

		prom, err = prometheus.NewConnection(testenv.PrometheusAddr)
		Expect(err).To(Succeed())

		esMonitor, err = monitor.NewConnection(testenv.MonitorAPIAddr)
		Expect(err).To(Succeed())
		
		// Returns true if there were at least three scrapes of a given metric
		threeScrapes := func(metric string) bool {
			return prom.HasData(fmt.Sprintf("count_over_time(%v[10m]) > 2", metric))
		}

		// Wait until there's some scrapes finished
		err = util.WaitUntilTrue(func() bool {
			return threeScrapes("kube_pod_info")
		}, "no kube_pod_info metric scrapes found")
		Expect(err).To(Succeed())
	})

var _ =	AfterSuite(func() {
	// Delete test app deployments + services
	for res := range appYamls {
		kube.DeleteYaml(args.ConsoleNamespace, res)

		// Ignore failures for now because deleting
		// 'es-test-service-with-only-endpoints' fails for some reason
		//Expect(err).To(Succeed())
	}

	testenv.CloseEnv()
})

var _ = Describe("all:prometheus", func() {
	// Helper for using metrics in table driven tests - metric name is both description and test function param
	Metric := func(name string) TableEntry {
		return Entry(name, name)
	}

	DescribeTable("basic metrics available",
		func(metric string) {
			Expect(prom.HasData(metric)).To(BeTrue())
		},
		Metric("prometheus_notifications_dropped_rate"),
		Metric("prometheus_notification_queue_percent"),
		Metric("prometheus_rule_evaluation_failures_rate"),
		Metric("prometheus_target_scrapes_exceeded_sample_limit_rate"),
		Metric("prometheus_tsdb_reloads_failures_rate"),
		Metric("prometheus_config_last_reload_successful"),
		Metric("prometheus_target_sync_percent"),
	)

	DescribeTable("health metrics available",
		func(metric string) {
			Expect(prom.HasData(fmt.Sprintf("model{name=\"%v\"}", metric))).To(BeTrue())
			Expect(prom.HasData(fmt.Sprintf("health{name=\"%v\"}", metric))).To(BeTrue())
		},
		Metric("prometheus_notifications_dropped"),
		Metric("prometheus_notification_queue"),
		Metric("prometheus_rule_evaluation_failures"),
		Metric("prometheus_target_too_many_metrics"),
		Metric("prometheus_tsdb_reloads_failures"),
		Metric("prometheus_target_down"),
		Metric("prometheus_config_reload_failed"),
		Metric("prometheus_scrape_time"),
	)

	It("coherency", func() {
		// Data with "es_workload" should also have a "namespace" label
		Expect(prom.HasData("count({es_workload=~\".+\", namespace=\"\", name!~\"node.*|kube_node.*\", __name__!~\"node.*|kube_node.*\"})")).To(BeFalse())
		// Health should have "es_workload" label, with a few known exceptions
		Expect(prom.HasData("health{es_workload=\"\", name!~\"node.*|kube_node.*|prometheus_target_down|scrape_time\"}")).To(BeFalse())

		// kube_pod_info must have es_workload labels (flaky!)
		//Expect(prom.HasData("kube_pd_info{es_workload=~\".+\"}")).To(BeTrue())

		Expect(prom.HasData("kube_pod_info{es_workload=\"\"}")).To(BeFalse())
		// kube data mapped pod to workload labels
		Expect(prom.HasData("{__name__=~ \"kube_.+\", pod!=\"\", es_workload=\"\"}")).To(BeFalse())
		// All container data have a workload label
		Expect(prom.HasData("{__name__=~\"container_.+\", es_workload=\"\"}")).To(BeFalse())
		// All targets should be reachable
		Expect(prom.HasData("up{kubernetes_name != \"es-test-service-with-only-endpoints\"} == 1")).To(BeTrue())
		Expect(prom.HasData("up{kubernetes_name != \"es-test-service-with-only-endpoints\"} == 0")).To(BeFalse())
	})

	DescribeTable("kube state metrics",
		func(metric string) {
			Expect(prom.HasData(metric)).To(BeTrue())
		},
		Metric("kube_pod_info"),
		Metric("kube_pod_ready"),
		Metric("kube_pod_container_restarts_rate"),
		Metric("kube_pod_failed"),
		Metric("kube_pod_not_running"),
		Metric("kube_workload_generation_lag"),
	)

	DescribeTable("kube state health",
		func(metric string) {
			Expect(prom.HasData(fmt.Sprintf("model{name=\"%v\"}", metric))).To(BeTrue())
			Expect(prom.HasData(fmt.Sprintf("health{name=\"%v\"}", metric))).To(BeTrue())
		},
		Metric("kube_container_restarts"),
		Metric("kube_pod_not_ready"),
		Metric("kube_pod_not_running"),
		Metric("kube_workload_generation_lag"),
	)

	// TODO: Check kube-state-metrics logs

	Context("app discovery", func() {
		It("'Pod' service", func() {
			appInstancesQuery := fmt.Sprintf("count( count by (instance) (ohai{es_workload=\"es-test\", namespace=\"%v\"}) ) == 2", args.ConsoleNamespace)
			err := util.WaitUntilTrue(func() bool {
				return prom.HasData(appInstancesQuery)
			}, "unable to discover app via pod service discovery")
			Expect(err).To(Succeed())

			// Pod discovery - automatic metrics should gain an es_monitor_type label when a custom monitor is created
			Expect(esMonitor.MakeMonitor("es-test/my_custom_monitor", "up")).To(Succeed())
			err = util.WaitUntilTrue(func() bool {
				return prom.HasModel("my_custom_monitor")
			}, "monitor my_custom_monitor wasn't created correctly")
			Expect(err).To(Succeed())
		})

		It("'Pod' service with multiple ports", func() {
			appInstancesWithMultiplePortsQuery := fmt.Sprintf("count( count by (instance) (ohai{es_workload=\"es-test-with-multiple-ports\", namespace=\"%v\"}) ) == 4", args.ConsoleNamespace)
			err := util.WaitUntilTrue(func() bool {
				return prom.HasData(appInstancesWithMultiplePortsQuery)
			}, "unable to discover app with multiple ports")
			Expect(err).To(Succeed())
		})

		It("'Service' discovery", func() {
			appInstancesViaServiceQuery := fmt.Sprintf("count( count by (instance) (ohai{es_workload=\"es-test-via-service\", namespace=\"%v\"}) ) == 2", args.ConsoleNamespace)
			err := util.WaitUntilTrue(func() bool {
				return prom.HasData(appInstancesViaServiceQuery)
			}, "unable to discover app via 'Service' service discovery")
			Expect(err).To(Succeed())

			// 'Service' discovery - automatic metrics should gain an es_monitor_type label when a custom monitor is created
			Expect(esMonitor.MakeMonitor("es-test-via-service/my_custom_monitor_for_service", "up")).To(Succeed())
			err = util.WaitUntilTrue(func() bool {
				return prom.HasModel("my_custom_monitor_for_service")
			}, "monitor my_custom_monitor_for_service wasn't created correctly")
			Expect(err).To(Succeed())
			Expect(prom.HasData("up{es_workload=\"es-test-via-service\", es_monitor_type=\"es-test-via-service\"}")).To(BeTrue())
		})

		It("'Service' endpoint discovery", func() {
			err := util.WaitUntilTrue(func() bool {
				return prom.HasData(fmt.Sprintf("count( count by (instance) (up{ "+
					"job=\"kubernetes-services\", kubernetes_name=\"es-test-service-with-only-endpoints\", namespace=\"%v\""+
					"}) ) == 1", args.ConsoleNamespace))
			}, "unable to discover app with only endpoints")
			Expect(err).To(Succeed())
		})

		It("kubernetes-cadvisor metrics", func() {
			// kubernetes-cadvisor metrics should have an es_monitor_type label.
			Expect(esMonitor.MakeMonitor("es-test/es-monitor-type-test", "container_cpu_load_average_10s")).To(Succeed())
			err := util.WaitUntilTrue(func() bool {
				return prom.HasModel("es-monitor-type-test")
			}, "monitor es-monitor-type-test wasn't created correctly")
			Expect(err).To(Succeed())
		})

		It("kubernetes-cadvisor regression", func() {
			// Specific test for regression of es-backend/issues/430.
			Expect(prom.HasData("{job=\"kubernetes-cadvisor\",es_monitor_type=\"es-test\"}")).To(BeTrue())
		})

		// The following is disabled due to flakyness
		XIt("metric data has es_monitor_type", func() {
			// Succeeds if all data for the workload es-test  has a matching es_monitor_type
			// Note we're currently ignoring health metrics because 'bad' data can stick around for 15m given their time window.
			err := util.WaitUntilTrue(func() bool {
				return prom.HasData("{es_workload=\"es-test\", es_monitor_type!=\"es-test\", __name__!=\"health\"}")
			}, "some data of es-test workload doesn't have matching es_monitor_type")
			Expect(err).To(Succeed())
		})
	})
})

func TestPrometheus(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Prometheus Suite")
}
