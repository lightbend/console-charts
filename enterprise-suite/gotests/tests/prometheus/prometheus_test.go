package prometheus

import (
	"fmt"
	"testing"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/args"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/testenv"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/kube"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/monitor"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/prometheus"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func TestPrometheus(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Prometheus Suite")
}

var prom *prometheus.Connection
var esMonitor *monitor.Connection

// Resource yaml file to deployment name map
var appYamls = map[string]string{
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

		// wait for deployment to become ready
		if len(depName) > 0 {
			err = util.WaitUntilSuccess(util.LongWait, func() error {
				return kube.IsDeploymentAvailable(testenv.K8sClient, args.ConsoleNamespace, depName)
			})
			Expect(err).To(Succeed())
		}
	}

	prom, err = prometheus.NewConnection(testenv.PrometheusAddr)
	Expect(err).To(Succeed())

	esMonitor, err = monitor.NewConnection(testenv.ConsoleAPIAddr)
	Expect(err).To(Succeed())

	waitForScrapes := func(metric string) error {
		return util.WaitUntilSuccess(util.LongWait, func() error {
			return prom.HasNScrapes(metric, 3)
		})
	}

	// wait until there's some scrapes finished
	Expect(waitForScrapes("prometheus_notifications_dropped_rate")).To(Succeed())
	Expect(waitForScrapes("model{name=\"prometheus_notifications_dropped\"}")).To(Succeed())
	Expect(waitForScrapes("kube_pod_info")).To(Succeed())
})

var _ = AfterSuite(func() {
	// Delete test app deployments + services
	for res := range appYamls {
		if err := kube.DeleteYaml(args.ConsoleNamespace, res); err != nil {
			util.LogG("Unable to remove %v\n", res)
		}
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
			Expect(prom.AnyData(metric)).To(Succeed())
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
			Expect(prom.AnyData(fmt.Sprintf("model{name=\"%v\"}", metric))).To(Succeed())
			Expect(prom.AnyData(fmt.Sprintf("health{name=\"%v\"}", metric))).To(Succeed())
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

	It("has the expected metrics", func() {
		// PromData with "es_workload" should also have a "namespace" label
		Expect(prom.HasNoData("count({es_workload=~\".+\", namespace=\"\", name!~\"node.*|kube_node.*\", __name__!~\"node.*|kube_node.*\"})")).To(Succeed())
		// Health should have "es_workload" label, with a few known exceptions
		Expect(prom.HasNoData("health{es_workload=\"\", name!~\"node.*|kube_node.*|prometheus_target_down|scrape_time\"}")).To(Succeed())

		// kube_pod_info must have es_workload labels (flaky!)
		//Expect(prom.HasData("kube_pd_info{es_workload=~\".+\"}")).To(BeTrue())

		Expect(prom.HasNoData("kube_pod_info{es_workload=\"\"}")).To(Succeed())
		// kube data mapped pod to workload labels
		Expect(prom.HasNoData("{__name__=~ \"kube_.+\", pod!=\"\", es_workload=\"\"}")).To(Succeed())
		// All container data have a workload label
		Expect(prom.HasNoData("{__name__=~\"container_.+\", es_workload=\"\"}")).To(Succeed())
		// All targets should be reachable
		Expect(prom.HasData("up{kubernetes_name != \"es-test-service-with-only-endpoints\"} == 1")).To(Succeed())
		Expect(prom.HasNoData(`up{kubernetes_name != "es-test-service-with-only-endpoints"} == 0`)).To(Succeed())
		// None of the metrics should have kubernetes_namespace label
		Expect(prom.HasNoData("{kubernetes_namespace!=\"\"}")).To(Succeed())
		// make sure the number of node_names matches the number of kubelets
		Expect(prom.HasData(`count (count by (node_name) ({node_name!="", job="kube-state-metrics"})) == count (kubelet_running_pod_count)`)).To(Succeed())
	})

	Context("kube-state-metrics", func() {
		It("should only scrape a single instance", func() {
			query := fmt.Sprintf(`count(kube_pod_status_ready{namespace="%s", es_workload="console-backend", condition="true"}) == 1`, args.ConsoleNamespace)
			Expect(prom.HasData(query)).To(Succeed())
		})
	})

	DescribeTable("kube state metrics",
		func(metric string) {
			Expect(prom.AnyData(metric)).To(Succeed())
		},
		Metric("kube_pod_info"),
		Metric("kube_pod_ready"),
		Metric("kube_pod_container_restarts_rate"),
		Metric("kube_pod_failed"),
	)

	DescribeTable("kube state health",
		func(metric string) {
			Expect(prom.AnyData(fmt.Sprintf("model{name=\"%v\"}", metric))).To(Succeed())
			Expect(prom.AnyData(fmt.Sprintf("health{name=\"%v\"}", metric))).To(Succeed())
		},
		Metric("kube_container_restarting"),
		Metric("kube_pod_not_ready"),
	)

	Context("k8s service discovery", func() {
		It("can discover a pod", func() {
			appInstancesQuery := fmt.Sprintf("count( count by (instance) (ohai{es_workload=\"es-test\", namespace=\"%v\"}) ) == 2", args.ConsoleNamespace)
			err := util.WaitUntilSuccess(util.LongWait, func() error {
				return prom.HasData(appInstancesQuery)
			})
			Expect(err).To(Succeed())
		})

		It("can create monitors on automatic metrics for Pods, aka `up`", func() {
			Expect(esMonitor.MakeSimpleMonitor("es-test/my_custom_monitor", "up")).To(Succeed())
			err := util.WaitUntilSuccess(util.LongWait, func() error {
				return prom.HasModel("my_custom_monitor")
			})
			Expect(err).To(Succeed())

			err = util.WaitUntilSuccess(util.LongWait, func() error {
				return prom.HasData(`up{es_workload="es-test", es_monitor_type="es-test"}`)
			})
			Expect(err).To(Succeed())

			Expect(esMonitor.DeleteMonitor("es-test/my_custom_monitor")).To(Succeed())
		})

		It("can discover a pod with multiple ports", func() {
			appInstancesWithMultiplePortsQuery := fmt.Sprintf("count( count by (instance) (ohai{es_workload=\"es-test-with-multiple-ports\", namespace=\"%v\"}) ) == 4", args.ConsoleNamespace)
			err := util.WaitUntilSuccess(util.LongWait, func() error {
				return prom.HasData(appInstancesWithMultiplePortsQuery)
			})
			Expect(err).To(Succeed())
		})

		It("can discover k8s `Service` resources", func() {
			appInstancesViaServiceQuery := fmt.Sprintf("count( count by (instance) (ohai{es_workload=\"es-test-via-service\", namespace=\"%v\"}) ) == 2", args.ConsoleNamespace)
			err := util.WaitUntilSuccess(util.LongWait, func() error {
				if err := prom.HasData(appInstancesViaServiceQuery); err != nil {
					return fmt.Errorf("unable to discover app via 'Service' service discovery: %v", err)
				}
				return nil
			})
			Expect(err).To(Succeed())
		})

		It("can create monitors on automatic metrics for Services, aka `up`", func() {
			// 'Service' discovery - automatic metrics should gain an es_monitor_type label when a custom monitor is created
			Expect(esMonitor.MakeSimpleMonitor("es-test-via-service/my_custom_monitor_for_service", "up")).To(Succeed())
			err := util.WaitUntilSuccess(util.LongWait, func() error {
				return prom.HasModel("my_custom_monitor_for_service")
			})
			Expect(err).To(Succeed())

			err = util.WaitUntilSuccess(util.LongWait, func() error {
				return prom.HasData("up{es_workload=\"es-test-via-service\", es_monitor_type=\"es-test-via-service\"}")
			})
			Expect(err).To(Succeed())

			Expect(esMonitor.DeleteMonitor("es-test-via-service/my_custom_monitor_for_service")).To(Succeed())
		})

		It("can discover Services without any pods, only endpoints, to support external redirection", func() {
			err := util.WaitUntilSuccess(util.LongWait, func() error {
				return prom.HasData(fmt.Sprintf("count( count by (instance) (up{ "+
					"job=\"kubernetes-services\", kubernetes_name=\"es-test-service-with-only-endpoints\", namespace=\"%v\""+
					"}) ) == 1", args.ConsoleNamespace))
			})
			Expect(err).To(Succeed())
		})

		It("kubelet cadvisor metrics should have a es_monitor_type label", func() {
			// kubernetes-cadvisor metrics should have an es_monitor_type label.
			Expect(esMonitor.MakeSimpleMonitor("es-test/es-monitor-type-test", "container_cpu_load_average_10s")).To(Succeed())
			err := util.WaitUntilSuccess(util.LongWait, func() error {
				return prom.HasModel("es-monitor-type-test")
			})
			Expect(err).To(Succeed())

			err = util.WaitUntilSuccess(util.LongWait, func() error {
				return prom.HasData(`{job="kubernetes-cadvisor", es_monitor_type="es-test"}`)
			})
			Expect(err).To(Succeed())

			Expect(esMonitor.DeleteMonitor("es-test/es-monitor-type-test")).To(Succeed())
		})

		XIt("metric data has es_monitor_type", func() {
			// Succeeds if all data for the workload es-test  has a matching es_monitor_type
			// Note we're currently ignoring health metrics because 'bad' data can stick around for 15m given their time window.
			err := util.WaitUntilSuccess(util.LongWait, func() error {
				if err := prom.HasData(`{es_workload="es-test", es_monitor_type!="es-test", __name__!="health"}`); err != nil {
					return fmt.Errorf("es-test workload data is missing es_monitor_type label: %v", err)
				}
				return nil
			})
			Expect(err).To(Succeed())
		})
	})
})
