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

const DefaultTimeout = 5 * time.Second

type CmdBuilder struct {
	name                 string
	args                 []string
	envVars              []string
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
		envVars:              nil,
		timeout:              DefaultTimeout,
		expectZeroExitStatus: true,
		panicOnError:         false,
		printStdout:          false,
		printStderr:          false,
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

func (cb *CmdBuilder) Env(name string, value string) *CmdBuilder {
	cb.envVars = append(cb.envVars, fmt.Sprintf("%v=%v", name, value))
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

	// Set up env variables
	if len(cb.envVars) > 0 {
		cmd.Env = os.Environ()
		for _, envVar := range cb.envVars {
			cmd.Env = append(cmd.Env, envVar)
		}
	}

	var exitcode int
	var cmderr error

	cmdstdout, err := cmd.StdoutPipe()
	if err != nil {
		panic("unable to get command stdout pipe")
	}

	cmdstderr, err := cmd.StderrPipe()
	if err != nil {
		panic("unable to get command stderr pipe")
	}

	// Run the command
	if err := cmd.Start(); err != nil {
		panic("unable to execute command")
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
		_, err = io.Copy(io.MultiWriter(stdoutWriters...), cmdstdout)
		if err != nil {
			panic("unable to copy stdout command pipe to io.MultiWriter")
		}

		_, err = io.Copy(io.MultiWriter(stderrWriters...), cmdstderr)
		if err != nil {
			panic("unable to copy stderr command pipe to io.MultiWriter")
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

	// Make an error on non-zero exit
	if cb.expectZeroExitStatus && cmderr == nil && exitcode != 0 {
		cmderr = fmt.Errorf("'%v %v' return status %v", cb.name, strings.Join(cb.args[:], " "), exitcode)
	}

	if cb.panicOnError && cmderr != nil {
		panic(cmderr)
	}

	return exitcode, cmderr
}

const maxRepeats = 30
const firstSleepMs = 20
const maxSleepMs = 10000

// Repeatedly runs a function, sleeping for a bit after each time, until it returns true or reaches maxRepeats.
// failMsg will be printed in the result error in case of timeout, can be empty.
func WaitUntilTrue(f func() bool, failMsg string) error {
	sleepTimeMs := firstSleepMs
	for i := 0; i < maxRepeats; i++ {
		if f() {
			return nil
		}
		// Exponential backoff
		sleepTimeMs = sleepTimeMs * 2
		if sleepTimeMs > maxSleepMs {
			sleepTimeMs = maxSleepMs
		}
		time.Sleep(time.Duration(sleepTimeMs) * time.Millisecond)
	}

	timeoutMsg := "WaitUntilTrue reached maximum repeats without f() returning true"
	if failMsg != "" {
		return fmt.Errorf("%s: %s", timeoutMsg, failMsg)
	}
	return errors.New(timeoutMsg)
}

// Repeatedly runs a function, sleeping for a bit after each time, until it succeeds or reaches maxRepeats
func WaitUntilSuccess(f func() error) error {
	var lastErr error
	if WaitUntilTrue(func() bool {
		lastErr = f()
		return lastErr == nil
	}, "") != nil {
		return fmt.Errorf("WaitUntilSuccess reached maximum repeats without f() succeeding: %v", lastErr)
	}
	return nil
}
