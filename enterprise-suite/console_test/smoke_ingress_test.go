package console

import (
	"fmt"
	"net/http"
	"time"

	extv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"

	"github.com/lightbend/console_test/args"

	"github.com/lightbend/console_test/util/minikube"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// Console NodePort service
const consoleServiceName = "expose-es-console"

// Name of Ingress that this test creates
const testIngressName = "console-test-ingress"

// TODO(mitkus): Make this test work. On my machine both this and bash e2e smoke_ingress
// stopped working after minikube update.
var _ = XDescribe("minikube:ingress", func() {
	It("responds to requests", func() {
		// On repeated test runs the ingress might already exist, so check before creating a new one
		_, err := k8sClient.ExtensionsV1beta1().Ingresses(args.ConsoleNamespace).Get(testIngressName, metav1.GetOptions{})
		if err != nil {

			// Figure out which port expose-es-console service uses
			consoleService, err := k8sClient.CoreV1().Services(args.ConsoleNamespace).Get(consoleServiceName, metav1.GetOptions{})
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
			_, err = k8sClient.ExtensionsV1beta1().Ingresses(args.ConsoleNamespace).Create(ingress)
			Expect(err).To(Succeed())

			// TODO: Proper waiting here
			time.Sleep(time.Minute)
		}

		ip, err := minikube.Ip()
		Expect(err).To(Succeed())

		httpClient := &http.Client{}
		req, err := http.NewRequest("GET", fmt.Sprintf("http://%v/es-console", ip), nil)
		Expect(err).To(Succeed())

		req.Host = "minikube.ingress.test"
		resp, err := httpClient.Do(req)
		Expect(err).To(Succeed())
		Expect(resp.StatusCode).To(Equal(200))
	})
})
