package ingress

import (
	"fmt"
	"net/http"
	"testing"

	extv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"

	"github.com/lightbend/go_tests/args"
	"github.com/lightbend/go_tests/testenv"

	"github.com/lightbend/go_tests/util"
	"github.com/lightbend/go_tests/util/minikube"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// Console NodePort service
const consoleServiceName = "expose-es-console"

// Name of Ingress that this test creates
const testIngressName = "console-test-ingress"

var _ = BeforeSuite(func() {
	testenv.InitEnv()
})

var _ = AfterSuite(func() {
	testenv.CloseEnv()
})

var _ = Describe("minikube:ingress", func() {
	It("responds to requests", func() {
		coreAPI := testenv.K8sClient.CoreV1()
		extAPI := testenv.K8sClient.ExtensionsV1beta1()

		// On repeated test runs the ingress might already exist, so check before creating a new one
		_, err := extAPI.Ingresses(args.ConsoleNamespace).Get(testIngressName, metav1.GetOptions{})
		if err != nil {

			// Figure out which port expose-es-console service uses
			consoleService, err := coreAPI.Services(args.ConsoleNamespace).Get(consoleServiceName, metav1.GetOptions{})
			Expect(err).To(Succeed())
			servicePort := consoleService.Spec.Ports[0].Port

			// Create ingress
			ingress := &extv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: testIngressName,
					Annotations: map[string]string{
						"nginx.ingress.kubernetes.io/rewrite-target": "/",
					},
				},
				Spec: extv1.IngressSpec{
					Rules: []extv1.IngressRule{
						extv1.IngressRule{
							"minikube.ingress.test",
							extv1.IngressRuleValue{
								&extv1.HTTPIngressRuleValue{
									Paths: []extv1.HTTPIngressPath{
										extv1.HTTPIngressPath{
											Path: "/es-console",
											Backend: extv1.IngressBackend{
												ServiceName: consoleServiceName,
												ServicePort: intstr.FromInt(int(servicePort)),
											},
										},
									},
								},
							},
						},
					},
				},
			}
			_, err = extAPI.Ingresses(args.ConsoleNamespace).Create(ingress)
			Expect(err).To(Succeed())
		}

		ip, err := minikube.Ip()
		Expect(err).To(Succeed())

		httpClient := &http.Client{}
		req, err := http.NewRequest("GET", fmt.Sprintf("http://%v/es-console", ip), nil)
		Expect(err).To(Succeed())

		req.Host = "minikube.ingress.test"

		err = util.WaitUntilTrue(func() bool {
			resp, err := httpClient.Do(req)
			return err == nil && resp.StatusCode == 200
		}, "console not accessible through minikube ingress")
		Expect(err).To(Succeed())
	})
})

func TestIngress(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ingress Suite")
}