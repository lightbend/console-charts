package args

import (
	"flag"
	"os"
	"path/filepath"
)

// Kubeconfig is an absolute path to the kubeconfig file
var Kubeconfig string

// ConsoleNamespace is name for kubernetes namespace where tests will run
var ConsoleNamespace string

func init() {
	homeDir := os.Getenv("HOME")

	flag.StringVar(&Kubeconfig, "kubeconfig", filepath.Join(homeDir, ".kube", "config"), "absolute path to the kubeconfig file")
	flag.StringVar(&ConsoleNamespace, "namespace", "lightbend-test", "kubernetes namespace where tests will run")
}
