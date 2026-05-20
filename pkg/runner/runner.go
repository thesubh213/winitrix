package runner

import (
	"bufio"
	"bytes"
	"context"
	"os/exec"
	"sync"

	"github.com/thesubh213/winitrix/pkg/logger"
)

// Result contains the output and error of a command.
type Result struct {
	Stdout string
	Stderr string
	Err    error
}

// ProgressCallback is called with each line of output from a command.
// It allows streaming progress updates during command execution.
type ProgressCallback func(line string)

// StreamResult contains the result of a streamed command execution.
type StreamResult struct {
	Stdout string
	Stderr string
	Err    error
	Cmd    *exec.Cmd // Reference to the command for potential signal handling
}

// RunSilent executes a command without printing to the console.
// It hides the window on Windows.
func RunSilent(ctx context.Context, name string, args ...string) Result {
	logger.Debug("Executing: %s %v", name, args)

	cmd := exec.CommandContext(ctx, name, args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	setSysProcAttr(cmd)

	err := cmd.Run()

	res := Result{
		Stdout: truncateOutput(stdoutBuf.String(), 10000), // Prevent memory bloat
		Stderr: truncateOutput(stderrBuf.String(), 10000),
		Err:    err,
	}

	if ctx.Err() == context.DeadlineExceeded {
		res.Err = context.DeadlineExceeded
		logger.Error("Command timed out: %s %v", name, args)
	} else if err != nil {
		logger.Error("Command failed: %s %v\nError: %v\nStderr: %s", name, args, err, res.Stderr)
	} else {
		logger.Debug("Command succeeded: %s %v", name, args)
	}

	return res
}

func truncateOutput(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[len(s)-maxLen:]
	}
	return s
}

// RunWithProgress executes a command and streams output to a callback.
// This allows real-time monitoring of command progress.
func RunWithProgress(ctx context.Context, progressCb ProgressCallback, name string, args ...string) StreamResult {
	logger.Debug("Executing with progress: %s %v", name, args)

	cmd := exec.CommandContext(ctx, name, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return StreamResult{
			Err: err,
			Cmd: cmd,
		}
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return StreamResult{
			Err: err,
			Cmd: cmd,
		}
	}

	setSysProcAttr(cmd)

	err = cmd.Start()
	if err != nil {
		return StreamResult{
			Err: err,
			Cmd: cmd,
		}
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(2)

	// Scan stdout line by line and send to callback
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		scanner.Split(splitProgressLines)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			stdoutBuf.WriteString(line + "\n")
			if progressCb != nil {
				progressCb(line)
			}
		}
	}()

	// Capture stderr
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		scanner.Split(splitProgressLines)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			stderrBuf.WriteString(line + "\n")
			if progressCb != nil {
				progressCb(line)
			}
		}
	}()

	execErr := cmd.Wait()
	wg.Wait()

	res := StreamResult{
		Stdout: truncateOutput(stdoutBuf.String(), 10000),
		Stderr: truncateOutput(stderrBuf.String(), 10000),
		Err:    execErr,
		Cmd:    cmd,
	}

	if ctx.Err() == context.DeadlineExceeded {
		res.Err = context.DeadlineExceeded
		logger.Error("Command timed out: %s %v", name, args)
	} else if execErr != nil {
		logger.Error("Command failed: %s %v\nError: %v\nStderr: %s", name, args, execErr, res.Stderr)
	} else {
		logger.Debug("Command succeeded: %s %v", name, args)
	}

	return res
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
