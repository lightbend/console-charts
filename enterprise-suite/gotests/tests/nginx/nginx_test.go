// nginx tests the nginx config rules work as expected.
package nginx

import (
	"fmt"
	"strings"
	"testing"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/testenv"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/urls"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func TestNginx(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Nginx Suite")
}

var _ = BeforeSuite(func() {
	testenv.InitEnv()
})

var _ = AfterSuite(func() {
	testenv.CloseEnv()
})

var _ = Describe("all:nginx", func() {

	DescribeTable("services are accessible", func(service string) {
		addr := fmt.Sprintf("%v/%v", testenv.ConsoleAddr, service)
		By(addr)
		_, err := urls.Get200(addr)
		Expect(err).ToNot(HaveOccurred())

		basepath := "fee/fie/fou/fum"
		addr = fmt.Sprintf("%v/%v/%v", testenv.ConsoleAddr, basepath, service)
		By(addr)
		_, err = urls.Get200(addr)
		Expect(err).ToNot(HaveOccurred())
	},
		Entry("console", ""),
		Entry("grafana", "/service/grafana/"),
		Entry("prometheus", "/service/prometheus/"),
		Entry("console-api", "/service/console-api/status"),
	)

	DescribeTable("favicon.ico", func(base string) {
		addr := fmt.Sprintf("%v/%vfavicon.ico", testenv.ConsoleAddr, base)
		By(addr)
		res, err := urls.Get200(addr)
		Expect(err).ToNot(HaveOccurred())
		Expect(res.Headers.Get("Content-Type")).To(HavePrefix("image/"))
	},
		Entry("root", ""),
		Entry("prefix", "/my/monitoring/"),
	)

	DescribeTable("is redirected", func(service *string, location string) {
		By(*service)
		resp, err := urls.Get(*service, false)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Status).To(Equal(301), "status code should be a redirect")
		Expect(resp.Headers.Get("Location")).To(Equal(location))
	},
		Entry("prometheus", &testenv.PrometheusAddr, "/service/prometheus/"),
		Entry("grafana", &testenv.GrafanaAddr, "/service/grafana/"),
	)

	DescribeTable("caching is off", func(addr string) {
		By(addr)
		cacheoff := "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0"
		resp, err := urls.Get200(testenv.ConsoleAddr + addr)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Headers.Get("Cache-Control")).To(Equal(cacheoff))
	},
		Entry("cluster page", "/cluster"),
		Entry("workload page", "/workloads/prometheus-server"),
		Entry("grafana frontend scripts", "/service/grafana/dashboard/script/exporter-async.js?service-type=cluster"),
	)

	DescribeTable("security headers are present", func(addr string) {
		By(addr)
		resp, err := urls.Get200(testenv.ConsoleAddr + addr)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Headers.Get("Server")).To(Equal(""))
		Expect(resp.Headers.Get("X-Frame-Options")).To(Equal("DENY"))
		Expect(resp.Headers.Get("X-XSS-Protection")).To(Equal("1"))
	},
		Entry("cluster", "/cluster"),
		Entry("workload", "/workloads/prometheus-server"),
		Entry("prometheus", "/service/prometheus/"),
		Entry("grafana", "/service/grafana/"),
		// es-monitor-api fails DENY check
		XEntry("console-api", "/service/es-monitor-api/"),
	)

	DescribeTable("csp headers are present", func(addr string) {
		By(addr)
		csp := "img-src 'self' data:; default-src 'self' 'unsafe-eval' 'unsafe-inline';"
		resp, err := urls.Get200(testenv.ConsoleAddr + addr)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Headers.Get("Content-Security-Policy")).To(Equal(csp))
	},
		Entry("cluster", "/cluster"),
		Entry("workload", "/workloads/prometheus-server"),
	)

	Context("xss guard", func() {
		It("guards against html injection", func() {
			resp, err := urls.Get200(testenv.ConsoleAddr + "/cluster%22/%3E%3Cscript%3Ealert(789)%3C/script%3E")
			Expect(err).ToNot(HaveOccurred())
			Expect(strings.Count(resp.Body, "<script>alert(789)</script>")).To(Equal(0))
		})
	})

	DescribeTable("should rewrite a missing trailing slash", func(prefix string) {
		url := testenv.ConsoleAddr + prefix
		By(url)
		res, err := urls.Get200(url)
		Expect(err).ToNot(HaveOccurred())
		Expect(res.Body).To(ContainSubstring(`<base href="%v/">`, prefix))
	},
		Entry("single subpath", "/monitoring"),
		Entry("multiple subpaths", "/my/monitoring/prefix"),
	)
})
