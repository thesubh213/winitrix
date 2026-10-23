package runner

import (
	"bufio"
	"context"
	"errors"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCappedBufferWriteString(t *testing.T) {
	b := newCappedBuffer(10)
	if n, err := b.WriteString("hello"); err != nil || n != 5 {
		t.Fatalf("WriteString returned (%d, %v)", n, err)
	}
	if n, err := b.WriteString("world!!!"); err != nil || n != 8 {
		t.Fatalf("WriteString must consume all input: (%d, %v)", n, err)
	}
	if got := b.String(); got != "helloworld" {
		t.Fatalf("buffer = %q, want %q", got, "helloworld")
	}
	if b.dropped != 3 {
		t.Fatalf("dropped = %d, want 3", b.dropped)
	}
}

func TestCappedBufferWrite(t *testing.T) {
	b := newCappedBuffer(4)
	if n, err := b.Write([]byte("abcdef")); err != nil || n != 6 {
		t.Fatalf("Write must consume all input: (%d, %v)", n, err)
	}
	if got := b.String(); got != "abcd" {
		t.Fatalf("buffer = %q, want %q", got, "abcd")
	}
	if b.dropped != 2 {
		t.Fatalf("dropped = %d, want 2", b.dropped)
	}

	// Writing once the cap is reached drops everything.
	_, _ = b.WriteString("more")
	if b.dropped != 6 {
		t.Fatalf("dropped = %d, want 6", b.dropped)
	}
}

func TestCappedBufferWriteByte(t *testing.T) {
	b := newCappedBuffer(2)
	for _, ch := range []byte("abc") {
		if err := b.WriteByte(ch); err != nil {
			t.Fatal(err)
		}
	}
	if b.String() != "ab" || b.dropped != 1 {
		t.Fatalf("got %q dropped=%d", b.String(), b.dropped)
	}
}

func TestSplitProgressLines(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"lf", "a\nb\n", []string{"a", "b"}},
		{"crlf", "a\r\nb\r\n", []string{"a", "b"}},
		{"cr only", "a\rb\r", []string{"a", "b"}},
		{"no trailing newline", "abc", []string{"abc"}},
		{"mixed", "a\r\nb\rc\n", []string{"a", "b", "c"}},
		{"empty", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := bufio.NewScanner(strings.NewReader(tt.input))
			scanner.Split(splitProgressLines)
			var got []string
			for scanner.Scan() {
				got = append(got, scanner.Text())
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExcerpt(t *testing.T) {
	small := strings.Repeat("x", 100)
	if got := excerpt(small); got != small {
		t.Fatalf("short strings must pass through untouched")
	}

	big := strings.Repeat("a", 3*logExcerptCap)
	got := excerpt(big)
	if len(got) >= len(big) {
		t.Fatalf("excerpt did not shrink output: %d >= %d", len(got), len(big))
	}
	if !strings.Contains(got, "truncated") {
		t.Fatalf("excerpt missing truncation marker: %q", got)
	}
}

func TestRunSilentCapturesOutput(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only command")
	}

	res := RunSilent(context.Background(), "cmd", "/c", "echo hello")
	if res.Err != nil {
		t.Fatalf("RunSilent failed: %v", res.Err)
	}
	if !strings.Contains(res.Stdout, "hello") {
		t.Fatalf("stdout = %q, want it to contain hello", res.Stdout)
	}
}

// TestRunWithProgressCapturesTail guards the exec pipe-draining order: Wait
// must come after both scanners finish, otherwise the tail of the output is
// silently dropped.
func TestRunWithProgressCapturesTail(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only command")
	}

	script := "for /L %i in (1,1,2000) do @echo line %i & echo TAIL_MARKER"
	res := RunWithProgress(context.Background(), nil, "cmd", "/c", script)
	if res.Err != nil {
		t.Fatalf("RunWithProgress failed: %v", res.Err)
	}
	if !strings.Contains(res.Stdout, "TAIL_MARKER") {
		t.Fatalf("output tail was lost (%d bytes captured)", len(res.Stdout))
	}
	if !strings.Contains(res.Stdout, "line 1") {
		t.Fatalf("output head was lost")
	}
}

func TestRunWithProgressInvokesCallback(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only command")
	}

	var lines []string
	res := RunWithProgress(context.Background(), func(line string) {
		lines = append(lines, line)
	}, "cmd", "/c", "echo alpha & echo beta & echo gamma")
	if res.Err != nil {
		t.Fatalf("RunWithProgress failed: %v", res.Err)
	}
	if len(lines) < 3 {
		t.Fatalf("callback saw %d lines, want at least 3", len(lines))
	}
}

func TestRunSilentTimeoutIsReported(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only command")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel()

	res := RunSilent(ctx, "cmd", "/c", "ping -n 6 127.0.0.1 > nul")
	if !errors.Is(res.Err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", res.Err)
	}
}
