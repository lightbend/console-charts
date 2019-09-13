package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/onsi/ginkgo"
)

const DefaultTimeout = 15 * time.Second

type CmdBuilder struct {
	name                 string
	args                 []string
	envVars              []string
	timeout              time.Duration
	expectZeroExitStatus bool
	printStdout          bool
	printStderr          bool
	printCommand         bool
	cmd                  *exec.Cmd
	cancelFunc           context.CancelFunc
	cmdStdout            io.ReadCloser
	cmdStderr            io.ReadCloser
	captureStdout        *strings.Builder
	captureStderr        *strings.Builder
}

func Cmd(name string, args ...string) *CmdBuilder {
	return &CmdBuilder{
		name:         name,
		args:         args,
		envVars:      nil,
		timeout:      DefaultTimeout,
		printStdout:  false,
		printStderr:  false,
		printCommand: false,
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

// PrintCommand will cause the command and its arguments to be printed before it executes.
func (cb *CmdBuilder) PrintCommand() *CmdBuilder {
	cb.printCommand = true
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

	if cb.printCommand {
		LogG("%v\n", cb)
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
	cb.cmd.Env = os.Environ()
	for _, envVar := range cb.envVars {
		cb.cmd.Env = append(cb.cmd.Env, envVar)
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

	go func() {
		if _, err := io.Copy(io.MultiWriter(stdoutWriters...), cb.cmdStdout); err != nil {
			fmt.Fprintf(ginkgo.GinkgoWriter, "%v: unable to copy process stdout: %v", cb, err)
		}
	}()

	go func() {
		if _, err := io.Copy(io.MultiWriter(stderrWriters...), cb.cmdStderr); err != nil {
			fmt.Fprintf(ginkgo.GinkgoWriter, "%v: unable to copy process stderr: %v", cb, err)
		}
	}()

	// Run the command
	if err := cb.cmd.Start(); err != nil {
		// If command is unavailable on the system we end up here
		return fmt.Errorf("%v: unable to execute: %v", cb, err)
	}

	return nil
}

func (cb *CmdBuilder) wait(ignoreExitErr bool) error {
	if cb == nil || cb.cmd == nil {
		return fmt.Errorf("%v: tried to Wait() for a command that was never started", cb)
	}
	// Ensure we clean up the golang command context.
	defer cb.cancelFunc()

	if err := cb.cmd.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); ok && ignoreExitErr {
			// We don't care as long as it exited.
			return nil
		}
		return fmt.Errorf("%v: %v", cb, err)
	}
	return nil
}

// Run starts and then waits for a process to finish.
func (cb *CmdBuilder) Run() error {
	if err := cb.start(); err != nil {
		return err
	}
	return cb.wait(false)
}

// StartAsync starts the process without waiting for its results. It will check it hasn't immediately died.
// Use StopAsync() to stop and wait for the results.
func (cb *CmdBuilder) StartAsync() error {
	// We should never timeout async processes.
	cb.NoTimeout()
	if err := cb.start(); err != nil {
		return nil
	}
	// Check that process hasn't immediately died.
	time.Sleep(100 * time.Millisecond)
	if cb.cmd.ProcessState != nil && cb.cmd.ProcessState.Exited() {
		return cb.wait(false)
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
	if err := cb.wait(true); err != nil {
		return fmt.Errorf("%v: %v", cb, err)
	}
	return nil
}

type WaitTime time.Duration

const (
	firstSleep = 10 * time.Millisecond
	maxSleep   = 10 * time.Second

	// Use for operations which are expected to succeed quickly.
	SmallWait = WaitTime(5 * time.Second)
	// Use for operators that can take a little while to start working.
	MedWait = WaitTime(15 * time.Second)
	// Use for operations which can take a while to succeed.
	LongWait = WaitTime(70 * time.Second)
	// Should not be used! Might be useful during development.
	LongestWait = WaitTime(200 * time.Second)
)

// Repeatedly runs a function, sleeping for a bit after each time, until it returns nil or reaches maxRepeats.
func WaitUntilSuccess(maxWait WaitTime, f func() error) error {
	sleepTime := firstSleep
	var lastErr error
	for endTime := time.Now().Add(time.Duration(maxWait)); time.Now().Before(endTime); {
		if lastErr = f(); lastErr == nil {
			return nil
		}
		// Exponential backoff
		sleepTime *= 2
		if sleepTime > maxSleep {
			sleepTime = maxSleep
		}
		time.Sleep(sleepTime)
	}

	return fmt.Errorf("WaitUntilSuccess failed: %v", lastErr)

}

func Close(closer io.Closer) {
	if err := closer.Close(); err != nil {
		// we never expect close to fail
		panic(err)
	}
}

// Returns a free TCP port on the local machine or an error
func FindFreePort() int {
	conn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(fmt.Sprintf("unexpecte error when trying to find free port: %v", err))
	}
	defer conn.Close()
	return conn.Addr().(*net.TCPAddr).Port
}

// LogG logs to the ginkgo writer. On test failure, it will print out whatever was written to it.
func LogG(format string, a ...interface{}) {
	if _, err := fmt.Fprintf(ginkgo.GinkgoWriter, format, a...); err != nil {
		panic(err)
	}
}

func IndentJson(in string) string {
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(in), "", "  "); err != nil {
		panic(err)
	}
	return buf.String()
}
