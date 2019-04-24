package installer

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/args"
	"github.com/lightbend/console-charts/enterprise-suite/gotests/util/lbc"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/testenv"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestInstaller(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Installer (lbc.py) Suite")
}

var _ = BeforeSuite(func() {
	testenv.InitEnv()
})

var _ = AfterSuite(func() {
	testenv.CloseEnv()
})

func write(file *os.File, content string) {
	if _, err := file.Write([]byte(content)); err != nil {
		panic(err)
	}
}

var _ = Describe("all:lbc.py", func() {
	var (
		valuesFile *os.File
	)

	BeforeEach(func() {
		var err error
		valuesFile, err = ioutil.TempFile("", "values-*.yaml")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := os.Remove(valuesFile.Name())
		Expect(err).To(Succeed())
	})

	Context("upgrades", func() {
		Context("disable persistent volumes", func() {
			// Assumption is testenv.InitEnv() sets usePersistentVolumes=true

			BeforeEach(func() {
				write(valuesFile, `usePersistentVolumes: false`)
			})

			It("should fail if we don't provide --delete-pvcs", func() {
				Expect(lbc.Install([]string{}, []string{"-f " + valuesFile.Name()})).ToNot(Succeed())
			})

			It("should succeed if we provide --delete-pvcs", func() {
				Expect(lbc.Install([]string{"--delete-pvcs"}, []string{"-f " + valuesFile.Name()})).To(Succeed())
			})
		})
	})

	Context("arg parsing", func() {
		It("should fail if conflicting namespaces", func() {
			Expect(lbc.Install([]string{"--namespace=" + args.ConsoleNamespace},
				[]string{"--namespace=my-busted-namespace"})).ToNot(Succeed())
		})
	})
})
