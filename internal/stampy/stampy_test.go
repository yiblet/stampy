package stampy

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yiblet/stampy/internal/stampy/template"
)

func TestCreateIOWithDefaults(t *testing.T) {
	reader, writer, cleanup, err := createIO("", "")
	if err != nil {
		t.Fatalf("createIO returned error with defaults: %v", err)
	}
	t.Cleanup(func() {
		if err := cleanup(); err != nil {
			t.Fatalf("cleanup returned error: %v", err)
		}
	})

	if got, ok := reader.(*os.File); !ok || got != os.Stdin {
		t.Fatalf("expected reader to be os.Stdin, got %T", reader)
	}

	if got, ok := writer.(*os.File); !ok || got != os.Stdout {
		t.Fatalf("expected writer to be os.Stdout, got %T", writer)
	}
}

func TestCreateIOWithFiles(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.txt")
	outputPath := filepath.Join(dir, "output.txt")

	if err := os.WriteFile(inputPath, []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	reader, writer, cleanup, err := createIO(inputPath, outputPath)
	if err != nil {
		t.Fatalf("createIO returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := cleanup(); err != nil {
			t.Fatalf("cleanup returned error: %v", err)
		}
	})

	if _, ok := reader.(*os.File); !ok {
		t.Fatalf("expected reader to be *os.File, got %T", reader)
	}
	if _, ok := writer.(*os.File); !ok {
		t.Fatalf("expected writer to be *os.File, got %T", writer)
	}

	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected output file to be created, stat error: %v", err)
	}
}

func TestProcessLinesElapsedAndDelta(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := newFakeClock(base, base.Add(2*time.Second))

	tpl, err := template.Parse("{elapsed:.1f}s Δ{delta:.1f}s {}")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	input := strings.NewReader("first line\nsecond line\n")
	var output bytes.Buffer

	if err := processLines(input, &output, tpl, clock); err != nil {
		t.Fatalf("processLines returned error: %v", err)
	}

	lines := splitOutput(output.String())
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	if lines[0] != "0.0s Δ2.0s first line" {
		t.Fatalf("unexpected first line: %q", lines[0])
	}
	if lines[1] != "2.0s Δ0.0s second line" {
		t.Fatalf("unexpected second line: %q", lines[1])
	}
}

func TestProcessLinesRespectsMissingTrailingNewline(t *testing.T) {
	base := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	clock := newFakeClock(base)

	tpl, err := template.Parse("{elapsed:.1f}s {}")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	input := strings.NewReader("no newline")
	var output bytes.Buffer

	if err := processLines(input, &output, tpl, clock); err != nil {
		t.Fatalf("processLines returned error: %v", err)
	}

	result := output.String()
	if strings.HasSuffix(result, "\n") {
		t.Fatalf("expected output to omit trailing newline, got %q", result)
	}
}

func TestProcessLinesSingleLineWithNewline(t *testing.T) {
	base := time.Date(2024, 4, 1, 12, 0, 0, 0, time.UTC)
	clock := newFakeClock(base)

	tpl, err := template.Parse("{elapsed:.1f}s {}")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	input := strings.NewReader("only line\n")
	var output bytes.Buffer

	if err := processLines(input, &output, tpl, clock); err != nil {
		t.Fatalf("processLines returned error: %v", err)
	}

	if output.String() != "0.0s only line\n" {
		t.Fatalf("unexpected output: %q", output.String())
	}
}

func TestRunWithClockProcessesFileIO(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.txt")
	outputPath := filepath.Join(dir, "output.txt")

	if err := os.WriteFile(inputPath, []byte("example line\n"), 0o644); err != nil {
		t.Fatalf("failed to seed input file: %v", err)
	}

	stamp := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	clock := newFakeClock(stamp)

	opts := Options{Template: "{time:15:04}: {}", TemplateProvided: true, Input: inputPath, Output: outputPath}
	if err := RunWithClock(opts, clock); err != nil {
		t.Fatalf("RunWithClock returned error: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	got := string(data)
	if got != "12:00: example line\n" {
		t.Fatalf("unexpected output contents: %q", got)
	}
}

func TestRunWithClockDefaultsTemplate(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.txt")
	outputPath := filepath.Join(dir, "output.txt")

	if err := os.WriteFile(inputPath, []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("failed to seed input file: %v", err)
	}

	base := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	clock := newFakeClock(base)

	opts := Options{Input: inputPath, Output: outputPath}
	if err := RunWithClock(opts, clock); err != nil {
		t.Fatalf("RunWithClock returned error: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if string(data) != "0.0s hello\n" {
		t.Fatalf("unexpected default output: %q", string(data))
	}
}

func TestRunWithClockInvalidTemplate(t *testing.T) {
	stamp := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	clock := newFakeClock(stamp)

	if err := RunWithClock(Options{Template: "{elapsed", TemplateProvided: true}, clock); err == nil {
		t.Fatalf("expected parse error")
	}
}

func newFakeClock(times ...time.Time) func() time.Time {
	if len(times) == 0 {
		panic("newFakeClock requires at least one time value")
	}

	fc := &fakeClock{times: times}
	return fc.Now
}

type fakeClock struct {
	times []time.Time
	idx   int
}

func (f *fakeClock) Now() time.Time {
	if f.idx >= len(f.times) {
		return f.times[len(f.times)-1]
	}

	current := f.times[f.idx]
	f.idx++
	return current
}

func splitOutput(out string) []string {
	out = strings.TrimSuffix(out, "\n")
	if out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}

func splitLine(line string, t *testing.T, idx int) (string, string) {
	stamp, rest, ok := strings.Cut(line, ": ")
	if !ok {
		t.Fatalf("line %d missing separator: %q", idx, line)
	}
	return stamp, rest
}
