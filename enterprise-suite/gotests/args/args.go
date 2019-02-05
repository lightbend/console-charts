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

// TillerNamespace is the namespace where Helm Tiller is installed
var TillerNamespace string

// Minishift must be true if running on minishift
var Minishift bool

func init() {
	homeDir := os.Getenv("HOME")

	flag.StringVar(&Kubeconfig, "kubeconfig", filepath.Join(homeDir, ".kube", "config"), "absolute path to the kubeconfig file")
	flag.StringVar(&ConsoleNamespace, "namespace", "lightbend-test", "kubernetes namespace where tests will run")
	flag.StringVar(&TillerNamespace, "tiller-namespace", "", "kubernetes namespace where tiller is installed, leave default if kube-system")
	flag.BoolVar(&Minishift, "minishift", false, "must be true if running on minishift")
}
