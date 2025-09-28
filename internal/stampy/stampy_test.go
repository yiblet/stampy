package stampy

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
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

func TestProcessLinesSecondsFormat(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := newFakeClock(base, base, base.Add(2*time.Second))

	input := strings.NewReader("first line\nsecond line\n")
	var output bytes.Buffer

	if err := ProcessLines(input, &output, "s", clock); err != nil {
		t.Fatalf("processLines returned error: %v", err)
	}

	lines := splitOutput(output.String())
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	expectedSuffixes := []string{"first line", "second line"}
	for i, line := range lines {
		stamp, rest := splitLine(line, t, i)
		if rest != expectedSuffixes[i] {
			t.Fatalf("line %d body mismatch: got %q want %q", i, rest, expectedSuffixes[i])
		}

		if !strings.HasSuffix(stamp, "s") {
			t.Fatalf("line %d missing seconds suffix: %q", i, stamp)
		}

		seconds := strings.TrimSuffix(stamp, "s")
		if _, err := strconv.ParseFloat(seconds, 64); err != nil {
			t.Fatalf("line %d stamp not parseable as float seconds: %v", i, err)
		}
	}
}

func TestProcessLinesMillisecondsFormat(t *testing.T) {
	base := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	clock := newFakeClock(base, base.Add(1500*time.Millisecond))

	input := strings.NewReader("single line\n")
	var output bytes.Buffer

	if err := ProcessLines(input, &output, "ms", clock); err != nil {
		t.Fatalf("processLines returned error: %v", err)
	}

	lines := splitOutput(output.String())
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}

	stamp, rest := splitLine(lines[0], t, 0)
	if rest != "single line" {
		t.Fatalf("unexpected line body: %q", rest)
	}

	if stamp != "1500ms" {
		t.Fatalf("unexpected milliseconds stamp: %q", stamp)
	}
}

func TestProcessLinesCustomFormat(t *testing.T) {
	base := time.Date(2023, 12, 25, 18, 30, 0, 0, time.UTC)
	clock := newFakeClock(base, base)

	input := strings.NewReader("only line\n")
	var output bytes.Buffer

	layout := "2006-01-02"
	if err := ProcessLines(input, &output, layout, clock); err != nil {
		t.Fatalf("processLines returned error: %v", err)
	}

	lines := splitOutput(output.String())
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}

	stamp, rest := splitLine(lines[0], t, 0)
	if rest != "only line" {
		t.Fatalf("unexpected line body: %q", rest)
	}

	if _, err := time.Parse(layout, stamp); err != nil {
		t.Fatalf("timestamp does not match layout %q: %v", layout, err)
	}
}

func TestProcessLinesRespectsMissingTrailingNewline(t *testing.T) {
	base := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	clock := newFakeClock(base, base.Add(500*time.Millisecond))

	input := strings.NewReader("no newline")
	var output bytes.Buffer

	if err := ProcessLines(input, &output, "ms", clock); err != nil {
		t.Fatalf("processLines returned error: %v", err)
	}

	result := output.String()
	if strings.HasSuffix(result, "\n") {
		t.Fatalf("expected output to omit trailing newline, got %q", result)
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

	opts := Options{Format: "15:04", Input: inputPath, Output: outputPath}
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
