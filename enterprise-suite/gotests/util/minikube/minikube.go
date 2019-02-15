package minikube

import (
	"fmt"
	"strings"
	"time"

	"github.com/lightbend/gotests/util"
)

func IsRunning() bool {
	return util.Cmd("minikube", "status").Run() == nil
}

func Start(cpus int, mem int) error {
	err := util.Cmd("minikube", "start", fmt.Sprintf("--cpus=%v", cpus), fmt.Sprintf("--memory=%v", mem)).
		Timeout(time.Minute * 4).
		PrintOutput().
		Run()
	return err
}

func Stop() error {
	err := util.Cmd("minikube", "stop").
		Timeout(time.Minute).
		PrintOutput().
		Run()
	return err
}

func Delete() error {
	err := util.Cmd("minikube", "delete").
		Timeout(time.Minute).
		PrintOutput().
		Run()
	return err
}

func Ip() (string, error) {
	var result strings.Builder
	err := util.Cmd("minikube", "ip").
		CaptureStdout(&result).
		Run()

	ip := result.String()
	return strings.TrimRight(ip, "\n"), err
}
