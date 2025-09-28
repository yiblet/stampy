package internal

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yiblet/stampy/internal/template"
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

	if err := processLines(input, &output, tpl, Options{}, clock); err != nil {
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

	if err := processLines(input, &output, tpl, Options{}, clock); err != nil {
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

	if err := processLines(input, &output, tpl, Options{}, clock); err != nil {
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

	if string(data) != "2024-07-01T00:00:00Z: hello\n" {
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

func TestRunWithClockJSONLMode(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.jsonl")
	outputPath := filepath.Join(dir, "output.jsonl")

	// Create test input with mixed JSON content
	inputContent := `{"message": "hello", "level": "info"}
"just a string"
[1, 2, 3]
{invalid json
42
`
	if err := os.WriteFile(inputPath, []byte(inputContent), 0o644); err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	stamp := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	clock := newFakeClock(stamp, stamp.Add(1*time.Second), stamp.Add(2*time.Second),
		stamp.Add(3*time.Second), stamp.Add(4*time.Second))

	opts := Options{
		Template:         "{elapsed:.0f}s",
		TemplateProvided: true,
		Input:            inputPath,
		Output:           outputPath,
		JSONKey:          "timestamp",
	}

	if err := RunWithClock(opts, clock); err != nil {
		t.Fatalf("RunWithClock returned error: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	lines := splitOutput(string(data))
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d: %v", len(lines), lines)
	}

	// Test JSON object merge
	expected0 := `{"level":"info","message":"hello","timestamp":"0s"}`
	if lines[0] != expected0 {
		t.Errorf("line 0: expected %q, got %q", expected0, lines[0])
	}

	// Test primitive string wrap
	expected1 := `{"line":"just a string","timestamp":"1s"}`
	if lines[1] != expected1 {
		t.Errorf("line 1: expected %q, got %q", expected1, lines[1])
	}

	// Test array wrap
	expected2 := `{"line":[1,2,3],"timestamp":"2s"}`
	if lines[2] != expected2 {
		t.Errorf("line 2: expected %q, got %q", expected2, lines[2])
	}

	// Test invalid JSON wrap
	expected3 := `{"line":"{invalid json","timestamp":"3s"}`
	if lines[3] != expected3 {
		t.Errorf("line 3: expected %q, got %q", expected3, lines[3])
	}

	// Test number wrap
	expected4 := `{"line":42,"timestamp":"4s"}`
	if lines[4] != expected4 {
		t.Errorf("line 4: expected %q, got %q", expected4, lines[4])
	}
}

func TestProcessLinesJSONLModeWithComplexTemplate(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 30, 45, 0, time.UTC)
	clock := newFakeClock(base, base.Add(500*time.Millisecond))

	tpl, err := template.Parse("[{time:15:04:05}] {line} elapsed={elapsed:.2f}s")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	input := strings.NewReader(`{"user": "alice", "action": "login"}
{"user": "bob", "action": "logout"}
`)
	var output bytes.Buffer

	opts := Options{JSONKey: "event_time"}
	if err := processLines(input, &output, tpl, opts, clock); err != nil {
		t.Fatalf("processLines returned error: %v", err)
	}

	lines := splitOutput(output.String())
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	// Parse the JSON outputs to verify structure
	var line0, line1 map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &line0); err != nil {
		t.Fatalf("failed to parse line 0 as JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(lines[1]), &line1); err != nil {
		t.Fatalf("failed to parse line 1 as JSON: %v", err)
	}

	// Verify original JSON fields are preserved
	if line0["user"] != "alice" || line0["action"] != "login" {
		t.Errorf("line 0: original JSON fields not preserved: %v", line0)
	}
	if line1["user"] != "bob" || line1["action"] != "logout" {
		t.Errorf("line 1: original JSON fields not preserved: %v", line1)
	}

	// Verify event_time stamps are correctly formatted
	expectedStamp0 := "[12:30:45] 1 elapsed=0.00s"
	if line0["event_time"] != expectedStamp0 {
		t.Errorf("line 0: expected event_time %q, got %q", expectedStamp0, line0["event_time"])
	}

	expectedStamp1 := "[12:30:45] 2 elapsed=0.50s"
	if line1["event_time"] != expectedStamp1 {
		t.Errorf("line 1: expected event_time %q, got %q", expectedStamp1, line1["event_time"])
	}
}

func TestProcessLinesJSONLModePreservesNewlineHandling(t *testing.T) {
	base := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	clock := newFakeClock(base)

	tpl, err := template.Parse("{elapsed:.1f}s")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	// Input without trailing newline
	input := strings.NewReader(`"no newline"`)
	var output bytes.Buffer

	opts := Options{JSONKey: "ts"}
	if err := processLines(input, &output, tpl, opts, clock); err != nil {
		t.Fatalf("processLines returned error: %v", err)
	}

	result := output.String()
	if strings.HasSuffix(result, "\n") {
		t.Fatalf("expected output to omit trailing newline, got %q", result)
	}

	expected := `{"line":"no newline","ts":"0.0s"}`
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}
