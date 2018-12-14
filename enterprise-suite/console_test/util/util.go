package util

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

const DefaultTimeout = 3 * time.Second

type CmdBuilder struct {
	name                 string
	args                 []string
	timeout              time.Duration
	expectZeroExitStatus bool
	panicOnError         bool
	printStdout          bool
	printStderr          bool
}

func Cmd(name string, args ...string) *CmdBuilder {
	return &CmdBuilder{
		name:                 name,
		args:                 args,
		timeout:              DefaultTimeout,
		expectZeroExitStatus: true,
		panicOnError:         false,
	}
}

func (cb *CmdBuilder) Timeout(t time.Duration) *CmdBuilder {
	cb.timeout = t
	return cb
}

func (cb *CmdBuilder) NoTimeout() *CmdBuilder {
	cb.timeout = 0
	return cb
}

func (cb *CmdBuilder) AnyExitStatus() *CmdBuilder {
	cb.expectZeroExitStatus = false
	return cb
}

func (cb *CmdBuilder) PanicOnError() *CmdBuilder {
	cb.panicOnError = true
	return cb
}

func (cb *CmdBuilder) PrintStdout() *CmdBuilder {
	cb.printStdout = true
	return cb
}

func (cb *CmdBuilder) PrintStderr() *CmdBuilder {
	cb.printStderr = true
	return cb
}

func (cb *CmdBuilder) PrintOutput() *CmdBuilder {
	cb.printStdout = true
	cb.printStderr = true
	return cb
}

func (cb *CmdBuilder) Run() (int, error) {
	var cmd *exec.Cmd

	// Set up timeout context if needed
	if cb.timeout == 0 {
		cmd = exec.Command(cb.name, cb.args...)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), cb.timeout)
		cmd = exec.CommandContext(ctx, cb.name, cb.args...)
		defer cancel()
	}

	var exitcode int
	var cmderr error

	// Setup stdout/sterr pipes
	cmdstderr, err := cmd.StderrPipe()
	if err != nil {
		exitcode, cmderr = -100, err
	}
	cmdstdout, err := cmd.StdoutPipe()
	if err != nil {
		exitcode, cmderr = -200, err
	}

	if cmderr == nil {
		// Run the command
		if err := cmd.Start(); err != nil {
			// Arbitrary exit code for other errors
			exitcode, cmderr = -300, err
		} else {
			if cb.printStdout {
				io.Copy(os.Stdout, cmdstdout)
			}

			if cb.printStderr {
				io.Copy(os.Stderr, cmdstderr)
			}

			// Wait for command to finish
			if err := cmd.Wait(); err != nil {
				// Any non-zero return code results in an error, handle that here
				if exiterr, ok := err.(*exec.ExitError); ok {
					if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
						exitcode = status.ExitStatus()
					}
				} else {
					exitcode, cmderr = -400, err
				}
			}
		}
	}

	// Make an error on non-zero exit
	if cb.expectZeroExitStatus && cmderr == nil && exitcode != 0 {
		cmderr = errors.New(fmt.Sprintf("'%v %v' return status %v", cb.name, strings.Join(cb.args[:], " "), exitcode))
	}

	if cb.panicOnError && cmderr != nil {
		panic(cmderr)
	}

	return exitcode, cmderr
}
