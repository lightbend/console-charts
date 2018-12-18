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
	captureStdout        *strings.Builder
	captureStderr        *strings.Builder
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

func (cb *CmdBuilder) CaptureStdout(out *strings.Builder) *CmdBuilder {
	cb.captureStdout = out
	return cb
}

func (cb *CmdBuilder) CaptureStderr(out *strings.Builder) *CmdBuilder {
	cb.captureStderr = out
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
			var stdoutWriters []io.Writer
			var stderrWriters []io.Writer

			// Add string.Builder outputs
			if cb.captureStdout != nil {
				stdoutWriters = append(stdoutWriters, cb.captureStdout)
			}
			if cb.captureStderr != nil {
				stderrWriters = append(stderrWriters, cb.captureStderr)
			}

			// Add OS outputs
			if cb.printStdout {
				stdoutWriters = append(stdoutWriters, os.Stdout)
			}
			if cb.printStderr {
				stderrWriters = append(stderrWriters, os.Stderr)
			}

			// Hook up command stdout/stderr to potential destinations
			if len(stdoutWriters) > 0 {
				io.Copy(io.MultiWriter(stdoutWriters...), cmdstdout)
			}
			if len(stderrWriters) > 0 {
				io.Copy(io.MultiWriter(stderrWriters...), cmdstderr)
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
