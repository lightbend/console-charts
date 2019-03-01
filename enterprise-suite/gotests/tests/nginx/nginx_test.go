// nginx tests the nginx config rules work as expected.
package nginx

import (
	"fmt"
	"testing"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/args"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/testenv"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/urls"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func TestNginxRules(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Nginx Rules Suite")
}

var _ = BeforeSuite(func() {
	testenv.InitEnv()
})

var _ = AfterSuite(func() {
	testenv.CloseEnv()
})

var _ = Describe("all:nginx_rules", func() {
	DescribeTable("returns security headers for console endpoints", func(endpoint string) {
		url := testenv.ConsoleAddr + endpoint
		By(url)
		res, err := urls.Get200(url)
		Expect(err).ToNot(HaveOccurred())
		Expect(res.Headers).Should(HaveKeyWithValue("X-Frame-Options", []string{"DENY"}))
		Expect(res.Headers).Should(HaveKeyWithValue("X-Xss-Protection", []string{"1"}))
		Expect(res.Headers).Should(HaveKeyWithValue("Content-Security-Policy", []string{"img-src 'self' data:; default-src 'self' 'unsafe-eval' 'unsafe-inline';"}))
		Expect(res.Headers).Should(HaveKeyWithValue("Cache-Control", []string{"no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0"}))
	},
		Entry("cluster", "/cluster"),
		Entry("workloads", fmt.Sprintf("/namespaces/%v/workloads/tiller-deploy", args.TillerNamespace)),
		Entry("root", ""),
		Entry("prefix", "/monitoring"),
		// jsravn: Enable when pipelines is in console image.
		//Entry("pipelines", "/pipelines"),
	)

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
