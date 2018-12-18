package console

import (
	"fmt"
	"net/http"

	extv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"

	"github.com/lightbend/console_test/args"

	"github.com/lightbend/console_test/util/minikube"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("minikube:ingress", func() {
	It("is reachable using ingress", func() {
		consoleService, err := k8sClient.CoreV1().Services(args.ConsoleNamespace).Get("expose-es-console", metav1.GetOptions{})
		Expect(err).To(Succeed())
		servicePort := consoleService.Spec.Ports[0].Port

		ingress := &extv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-ingress",
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
											ServiceName: "expose-es-console",
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

		ip, err := minikube.Ip()
		Expect(err).To(Succeed())

		resp, err := http.Get(fmt.Sprintf("http://%v/es-console", ip))
		Expect(err).To(Succeed())
		Expect(resp.StatusCode).To(Equal(200))
	})
})
