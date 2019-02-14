package util

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/onsi/ginkgo"
)

const DefaultTimeout = 5 * time.Second

type CmdBuilder struct {
	name                 string
	args                 []string
	envVars              []string
	timeout              time.Duration
	expectZeroExitStatus bool
	printStdout          bool
	printStderr          bool
	cmd                  *exec.Cmd
	cancelFunc           context.CancelFunc
	cmdStdout            io.ReadCloser
	cmdStderr            io.ReadCloser
	captureStdout        *strings.Builder
	captureStderr        *strings.Builder
}

func Cmd(name string, args ...string) *CmdBuilder {
	return &CmdBuilder{
		name:        name,
		args:        args,
		envVars:     nil,
		timeout:     DefaultTimeout,
		printStdout: false,
		printStderr: false,
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

func (cb *CmdBuilder) String() string {
	return fmt.Sprintf("%v %v", cb.name, strings.Join(cb.args, " "))
}

func (cb *CmdBuilder) start() error {
	if cb.cmd != nil {
		panic(fmt.Sprintf("%v: attempted to start the same command multiple times", cb))
	}

	// Set up timeout context if needed
	if cb.timeout == 0 {
		cb.cmd = exec.Command(cb.name, cb.args...)
		cb.cancelFunc = func() {}
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), cb.timeout)
		cb.cmd = exec.CommandContext(ctx, cb.name, cb.args...)
		cb.cancelFunc = cancel
	}

	// Set up env variables
	if len(cb.envVars) > 0 {
		cb.cmd.Env = os.Environ()
		for _, envVar := range cb.envVars {
			cb.cmd.Env = append(cb.cmd.Env, envVar)
		}
	}

	cmdStdout, err := cb.cmd.StdoutPipe()
	if err != nil {
		panic(fmt.Sprintf("%v: unable to get command stdout pipe", cb))
	}

	cmdStderr, err := cb.cmd.StderrPipe()
	if err != nil {
		panic(fmt.Sprintf("%v: unable to get command stderr pipe", cb))
	}

	cb.cmdStdout = cmdStdout
	cb.cmdStderr = cmdStderr

	// Run the command
	if err := cb.cmd.Start(); err != nil {
		// If command is unavailable on the system we end up here
		return fmt.Errorf("%v: unable to execute: %v", cb, err)
	}

	return nil
}

func (cb *CmdBuilder) wait() error {
	if cb == nil || cb.cmd == nil {
		return fmt.Errorf("%v: tried to Wait() for a command that was never started", cb)
	}
	// Ensure we clean up the golang command context.
	defer cb.cancelFunc()

	// Copy stdout/stderr to potential destinations
	// Always print stdout/stderr to GinkgoWriter so that we can see all output in verbose mode and when test fails.
	stdoutWriters := []io.Writer{ginkgo.GinkgoWriter}
	stderrWriters := []io.Writer{ginkgo.GinkgoWriter}

	// Add external string.Builder outputs
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

	if _, err := io.Copy(io.MultiWriter(stdoutWriters...), cb.cmdStdout); err != nil {
		panic(fmt.Errorf("%v: unable to copy process stdout: %v", cb, err))
	}

	if _, err := io.Copy(io.MultiWriter(stderrWriters...), cb.cmdStderr); err != nil {
		panic(fmt.Errorf("%v: unable to copy process stderr: %v", cb, err))
	}

	if err := cb.cmd.Wait(); err != nil {
		return fmt.Errorf("%v: %v", cb, err)
	}
	return nil
}

// Run starts and then waits for a process to finish.
func (cb *CmdBuilder) Run() error {
	if err := cb.start(); err != nil {
		return err
	}
	return cb.wait()
}

// StartAsync starts the process without waiting for its results. It will check it hasn't immediately died.
// Use StopAsync() to stop and wait for the results.
func (cb *CmdBuilder) StartAsync() error {
	if err := cb.start(); err != nil {
		return nil
	}
	// Check that process hasn't immediately died
	time.Sleep(100 * time.Millisecond)
	if cb.cmd.ProcessState != nil && cb.cmd.ProcessState.Exited() {
		return cb.wait()
	}
	return nil
}

// StopAsync stops a process started with StartAsync(). It will send a SIGTERM and ensure the process stops.
// It will only return an error if the process fails to stop.
func (cb *CmdBuilder) StopAsync() error {
	if cb == nil || cb.cmd == nil {
		return fmt.Errorf("%v: can't stop what wasn't started", cb)
	}
	if err := cb.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		panic(fmt.Sprintf("%v: unable to send SIGTERM: %v", cb, err))
	}
	err := cb.wait()
	if _, ok := err.(*exec.ExitError); ok {
		// We don't care as long as it exited.
		return nil
	}
	return fmt.Errorf("%v: %v", cb, err)
}

const maxRepeats = 30
const firstSleepMs = 20
const maxSleepMs = 10000

// Repeatedly runs a function, sleeping for a bit after each time, until it returns nil or reaches maxRepeats.
func WaitUntilSuccess(f func() error) error {
	sleepTimeMs := firstSleepMs
	var lastErr error
	for i := 0; i < maxRepeats; i++ {
		if lastErr = f(); lastErr == nil {
			return nil
		}
		// Exponential backoff
		sleepTimeMs = sleepTimeMs * 2
		if sleepTimeMs > maxSleepMs {
			sleepTimeMs = maxSleepMs
		}
		time.Sleep(time.Duration(sleepTimeMs) * time.Millisecond)
	}

	return fmt.Errorf("WaitUntilSuccess reached maximum repeats without f() succeeding: %v", lastErr)
}

func Close(closer io.Closer) {
	if err := closer.Close(); err != nil {
		// we never expect close to fail
		panic(err)
	}
}
