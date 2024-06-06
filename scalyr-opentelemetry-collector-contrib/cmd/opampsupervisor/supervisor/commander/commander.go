// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package commander

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync/atomic"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/opampsupervisor/supervisor/config"
)

// Commander can start/stop/restart the Agent executable and also watch for a signal
// for the Agent process to finish.
type Commander struct {
	logger  *zap.Logger
	cfg     *config.Agent
	args    []string
	cmd     *exec.Cmd
	doneCh  chan struct{}
	running *atomic.Int64
}

func NewCommander(logger *zap.Logger, cfg *config.Agent, args ...string) (*Commander, error) {
	if cfg.Executable == "" {
		return nil, errors.New("agent.executable config option must be specified")
	}

	return &Commander{
		logger:  logger,
		cfg:     cfg,
		args:    args,
		running: &atomic.Int64{},
	}, nil
}

// Start the Agent and begin watching the process.
// Agent's stdout and stderr are written to a file.
// Calling this method when a command is already running
// is a no-op.
func (c *Commander) Start(ctx context.Context) error {
	if c.running.Load() == 1 {
		// Already started, nothing to do
		return nil
	}

	c.logger.Debug("Starting agent", zap.String("agent", c.cfg.Executable))

	logFilePath := "agent.log"
	logFile, err := os.Create(logFilePath)
	if err != nil {
		return fmt.Errorf("cannot create %s: %w", logFilePath, err)
	}

	c.cmd = exec.CommandContext(ctx, c.cfg.Executable, c.args...) // #nosec G204

	// Capture standard output and standard error.
	// https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/21072
	c.cmd.Stdout = logFile
	c.cmd.Stderr = logFile

	c.doneCh = make(chan struct{})

	if err := c.cmd.Start(); err != nil {
		return err
	}

	c.logger.Debug("Agent process started", zap.Int("pid", c.cmd.Process.Pid))
	c.running.Store(1)

	go c.watch()

	return nil
}

func (c *Commander) Restart(ctx context.Context) error {
	if err := c.Stop(ctx); err != nil {
		return err
	}

	return c.Start(ctx)
}

func (c *Commander) watch() {
	err := c.cmd.Wait()

	// cmd.Wait returns an exec.ExitError when the Collector exits unsuccessfully or stops
	// after receiving a signal. The Commander caller will handle these cases, so we filter
	// them out here.
	var exitError *exec.ExitError
	if ok := errors.As(err, &exitError); err != nil && !ok {
		c.logger.Error("An error occurred while watching the agent process", zap.Error(err))
	}

	c.running.Store(0)
	close(c.doneCh)
}

// Done returns a channel that will send a signal when the Agent process is finished.
func (c *Commander) Done() <-chan struct{} {
	return c.doneCh
}

// Pid returns Agent process PID if it is started or 0 if it is not.
func (c *Commander) Pid() int {
	if c.cmd == nil || c.cmd.Process == nil {
		return 0
	}
	return c.cmd.Process.Pid
}

// ExitCode returns Agent process exit code if it exited or 0 if it is not.
func (c *Commander) ExitCode() int {
	if c.cmd == nil || c.cmd.ProcessState == nil {
		return 0
	}
	return c.cmd.ProcessState.ExitCode()
}

func (c *Commander) IsRunning() bool {
	return c.running.Load() != 0
}

// Stop the Agent process. Sends SIGTERM to the process and wait for up 10 seconds
// and if the process does not finish kills it forcedly by sending SIGKILL.
// Returns after the process is terminated.
func (c *Commander) Stop(ctx context.Context) error {
	if c.running.Load() == 0 {
		// Not started, nothing to do.
		return nil
	}

	pid := c.cmd.Process.Pid
	c.logger.Debug("Stopping agent process", zap.Int("pid", pid))

	// Gracefully signal process to stop.
	if err := c.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return err
	}

	waitCtx, cancel := context.WithTimeout(ctx, 10*time.Second)

	// Setup a goroutine to wait a while for process to finish and send kill signal
	// to the process if it doesn't finish.
	var innerErr error
	go func() {
		<-waitCtx.Done()

		if !errors.Is(waitCtx.Err(), context.DeadlineExceeded) {
			c.logger.Debug("Agent process successfully stopped.", zap.Int("pid", pid))
			return
		}

		// Time is out. Kill the process.
		c.logger.Debug(
			"Agent process is not responding to SIGTERM. Sending SIGKILL to kill forcibly.",
			zap.Int("pid", pid))
		if innerErr = c.cmd.Process.Signal(syscall.SIGKILL); innerErr != nil {
			return
		}
	}()

	// Wait for process to terminate
	<-c.doneCh

	c.running.Store(0)

	// Let goroutine know process is finished.
	cancel()

	return innerErr
}
