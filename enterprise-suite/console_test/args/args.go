package args

import (
	"flag"
	"os"
	"path/filepath"
)

// StartMinikube will start a new minikube cluster if true
var StartMinikube bool

// Kubeconfig is an absolute path to the kubeconfig file
var Kubeconfig string

// Minikube is true if tests are running on minikube
var Minikube bool

// Openshift is true if tests are running on openshift
var Openshift bool

func init() {
	homeDir := os.Getenv("HOME")

	flag.BoolVar(&StartMinikube, "start-minikube", false, "if true, start minikube cluster instead of using existing cluster")
	flag.StringVar(&Kubeconfig, "kubeconfig", filepath.Join(homeDir, ".kube", "config"), "absolute path to the kubeconfig file")
	flag.BoolVar(&Minikube, "minikube", false, "run minikube specific tests")
	flag.BoolVar(&Openshift, "openshift", false, "run openshift specific tests")
}
