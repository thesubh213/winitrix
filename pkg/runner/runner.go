package runner

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/thesubh213/winitrix/pkg/logger"
)

// streamCap bounds how much of each output stream is retained in memory.
// Package managers can emit megabytes of progress refreshes; the cap keeps
// memory flat while still covering realistic output for counting results.
const streamCap = 2 << 20 // 2 MiB

// logExcerptCap bounds how much captured output lands in the log file.
const logExcerptCap = 4 << 10

// Result contains the output and error of a command.
type Result struct {
	Stdout  string
	Stderr  string
	Err     error
	Dropped int // bytes discarded after reaching the retention cap
}

// ProgressCallback is called with each line of output from a command.
// It allows streaming progress updates during command execution.
type ProgressCallback func(line string)

// StreamResult contains the result of a streamed command execution.
type StreamResult struct {
	Stdout  string
	Stderr  string
	Err     error
	Cmd     *exec.Cmd // Reference to the command for potential signal handling
	Dropped int
}

// cappedBuffer retains at most limit bytes and counts the remainder.
// Unlike a plain bytes.Buffer it cannot grow without bound.
type cappedBuffer struct {
	buf     bytes.Buffer
	limit   int
	dropped int
}

func newCappedBuffer(limit int) *cappedBuffer {
	return &cappedBuffer{limit: limit}
}

func (c *cappedBuffer) Write(p []byte) (int, error) {
	n := len(p)
	room := c.limit - c.buf.Len()
	switch {
	case room <= 0:
		c.dropped += n
	case n > room:
		_, _ = c.buf.Write(p[:room])
		c.dropped += n - room
	default:
		_, _ = c.buf.Write(p)
	}
	return n, nil
}

func (c *cappedBuffer) WriteString(s string) (int, error) {
	room := c.limit - c.buf.Len()
	switch {
	case room <= 0:
		c.dropped += len(s)
	case len(s) > room:
		_, _ = c.buf.WriteString(s[:room])
		c.dropped += len(s) - room
	default:
		_, _ = c.buf.WriteString(s)
	}
	return len(s), nil
}

func (c *cappedBuffer) WriteByte(b byte) error {
	if c.buf.Len() >= c.limit {
		c.dropped++
		return nil
	}
	return c.buf.WriteByte(b)
}

func (c *cappedBuffer) String() string { return c.buf.String() }

// newCommand builds a command whose cancellation kills the entire process
// tree. Package managers spawn children (git, aria2c, installers) that would
// otherwise survive a timeout or user cancellation.
func newCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	setSysProcAttr(cmd)
	cmd.WaitDelay = 5 * time.Second
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		return KillTree(cmd.Process.Pid)
	}
	return cmd
}

// RunSilent executes a command without printing to the console.
// It hides the window on Windows.
func RunSilent(ctx context.Context, name string, args ...string) Result {
	logger.Debug("Executing: %s %v", name, args)

	cmd := newCommand(ctx, name, args...)

	stdout := newCappedBuffer(streamCap)
	stderr := newCappedBuffer(streamCap)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()

	res := Result{
		Stdout:  stdout.String(),
		Stderr:  stderr.String(),
		Err:     err,
		Dropped: stdout.dropped + stderr.dropped,
	}

	switch {
	case ctx.Err() == context.DeadlineExceeded:
		res.Err = context.DeadlineExceeded
		logger.Error("Command timed out: %s %v", name, args)
	case ctx.Err() == context.Canceled:
		res.Err = context.Canceled
		logger.Info("Command cancelled: %s %v", name, args)
	case err != nil:
		logger.Error("Command failed: %s %v\nError: %v\nStderr: %s", name, args, err, excerpt(res.Stderr))
	default:
		logger.Debug("Command succeeded: %s %v", name, args)
	}

	return res
}

// RunWithProgress executes a command and streams output to a callback.
// This allows real-time monitoring of command progress.
func RunWithProgress(ctx context.Context, progressCb ProgressCallback, name string, args ...string) StreamResult {
	logger.Debug("Executing with progress: %s %v", name, args)

	cmd := newCommand(ctx, name, args...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return StreamResult{Err: err, Cmd: cmd}
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return StreamResult{Err: err, Cmd: cmd}
	}

	if err := cmd.Start(); err != nil {
		return StreamResult{Err: err, Cmd: cmd}
	}

	stdout := newCappedBuffer(streamCap)
	stderr := newCappedBuffer(streamCap)

	var wg sync.WaitGroup
	wg.Add(2)

	drain := func(r io.Reader, buf *cappedBuffer) {
		defer wg.Done()
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		scanner.Split(splitProgressLines)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			_, _ = buf.WriteString(line)
			_ = buf.WriteByte('\n')
			if progressCb != nil {
				progressCb(line)
			}
		}
	}

	go drain(stdoutPipe, stdout)
	go drain(stderrPipe, stderr)

	// Drain the pipes before calling Wait: Wait closes the pipe file
	// descriptors and would otherwise drop the tail of the output (where
	// "Successfully installed ..." summary lines live).
	wg.Wait()
	execErr := cmd.Wait()

	res := StreamResult{
		Stdout:  stdout.String(),
		Stderr:  stderr.String(),
		Err:     execErr,
		Cmd:     cmd,
		Dropped: stdout.dropped + stderr.dropped,
	}

	switch {
	case ctx.Err() == context.DeadlineExceeded:
		res.Err = context.DeadlineExceeded
		logger.Error("Command timed out: %s %v", name, args)
	case ctx.Err() == context.Canceled:
		res.Err = context.Canceled
		logger.Info("Command cancelled: %s %v", name, args)
	case execErr != nil:
		logger.Error("Command failed: %s %v\nError: %v\nStderr: %s", name, args, execErr, excerpt(res.Stderr))
	default:
		logger.Debug("Command succeeded: %s %v", name, args)
	}

	return res
}

// excerpt trims captured output for log lines (head + tail).
func excerpt(s string) string {
	if len(s) <= logExcerptCap {
		return s
	}
	half := logExcerptCap / 2
	return strings.ToValidUTF8(s[:half], "") +
		"\n... [truncated " + strconv.Itoa(len(s)-logExcerptCap) + " bytes] ...\n" +
		strings.ToValidUTF8(s[len(s)-half:], "")
}

func splitProgressLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		if data[i] == '\r' && i+1 < len(data) && data[i+1] == '\n' {
			return i + 2, dropTrailingCR(data[:i]), nil
		}
		return i + 1, dropTrailingCR(data[:i]), nil
	}

	if atEOF {
		return len(data), dropTrailingCR(data), nil
	}

	return 0, nil, nil
}

func dropTrailingCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[:len(data)-1]
	}
	return data
}
