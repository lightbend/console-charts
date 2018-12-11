package util

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

const DefaultTimeout = 3 * time.Second

// Executes a command with default timeout, returns exit status code or an error
func Exec(name string, arg ...string) (int, error) {
	return ExecWithTimeout(DefaultTimeout, name, arg...)
}

// Executes a command with given timeout, returns exit status code or an error
func ExecWithTimeout(timeout time.Duration, name string, arg ...string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := exec.CommandContext(ctx, name, arg...).Run(); err != nil {
		// Any non-zero return code is an error, handle that here
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), nil
			}
		}

		// Arbitrary exit code for other errors
		return -100, err
	}

	return 0, nil
}

// Executes a command with default timeout where exit code is expected to be zero, returns non-nil error otherwise
func ExecZeroExitCode(name string, arg ...string) error {
	return ExecZeroExitCodeWithTimeout(DefaultTimeout, name, arg...)
}

// Executes a command with given timeout where exit code is expected to be zero, returns non-nil error otherwise
func ExecZeroExitCodeWithTimeout(timeout time.Duration, name string, arg ...string) error {
	status, err := ExecWithTimeout(timeout, name, arg...)

	if err != nil {
		return err
	}

	if status != 0 {
		return errors.New(fmt.Sprintf("%v %v return status %v", name, strings.Join(arg[:], " "), status))
	}

	return nil
}
