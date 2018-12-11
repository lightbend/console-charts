package minikube

import (
	"fmt"
	"time"

	"github.com/lightbend/console_test/util"
)

func IsRunning() bool {
	if status, err := util.Exec("minikube", "status"); err != nil {
		return false
	} else {
		// minikube status exit code has bit encoded
		// 1 for minikube OK
		// 2 for cluster OK
		// 4 for kubernetes OK
		return status == (1 + 2 + 4)
	}
}

func Start(cpus int, mem int) error {
	return util.ExecZeroExitCodeWithTimeout(time.Minute,
		"minikube", "start", fmt.Sprintf("--cpus=%v", cpus), fmt.Sprintf("--memory=%v", mem))
}

func Stop() error {
	return util.ExecZeroExitCodeWithTimeout(time.Second*30, "minikube", "stop")
}

func Delete() error {
	return util.ExecZeroExitCodeWithTimeout(time.Second*30, "minikube", "delete")
}
