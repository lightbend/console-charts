package minikube

import (
	"fmt"
	"strings"
	"time"

	"github.com/lightbend/console_test/util"
)

func IsRunning() bool {
	if status, err := util.Cmd("minikube", "status").AnyExitStatus().Run(); err != nil {
		return false
	} else {
		return status == 0
	}
}

func Start(cpus int, mem int) error {
	_, err := util.Cmd("minikube", "start", fmt.Sprintf("--cpus=%v", cpus), fmt.Sprintf("--memory=%v", mem)).
		Timeout(time.Minute * 4).
		PrintOutput().
		Run()
	return err
}

func Stop() error {
	_, err := util.Cmd("minikube", "stop").
		Timeout(time.Minute).
		PrintOutput().
		Run()
	return err
}

func Delete() error {
	_, err := util.Cmd("minikube", "delete").
		Timeout(time.Minute).
		PrintOutput().
		Run()
	return err
}

func Ip() (string, error) {
	var result strings.Builder
	_, err := util.Cmd("minikube", "ip").
		CaptureStdout(&result).
		Run()

	ip := result.String()
	return strings.TrimRight(ip, "\n"), err
}
