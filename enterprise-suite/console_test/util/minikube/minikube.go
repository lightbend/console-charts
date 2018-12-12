package minikube

import (
	"fmt"
	"time"

	"github.com/lightbend/console_test/util"
)

func IsRunning() bool {
	if status, err := util.Cmd("minikube", "status").AnyExitStatus().Run(); err != nil {
		return false
	} else {
		// minikube status exit code has bit encoded state
		// 1 for minikube OK
		// 2 for cluster OK
		// 4 for kubernetes OK
		return status == (1 + 2 + 4)
	}
}

func Start(cpus int, mem int) error {
	_, err := util.Cmd("minikube", "start", fmt.Sprintf("--cpus=%v", cpus), fmt.Sprintf("--memory=%v", mem)).
		Timeout(time.Minute * 3).
		PrintStdout().
		PrintStderr().
		Run()
	return err
}

func Stop() error {
	_, err := util.Cmd("minikube", "stop").
		Timeout(time.Minute).
		PrintStdout().
		PrintStderr().
		Run()
	return err
}

func Delete() error {
	_, err := util.Cmd("minikube", "delete").
		Timeout(time.Minute).
		PrintStdout().
		PrintStderr().
		Run()
	return err
}
