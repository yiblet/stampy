package internal

import (
	"bytes"
	"testing"
	"time"

	"github.com/yiblet/stampy/internal/template"
)

func TestLineBufferPushAndFlush(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	later := start.Add(2 * time.Second)

	buffer := newLineBuffer()

	first := lineRecord{timestamp: start}
	if emit := buffer.push(first); emit != nil {
		t.Fatalf("expected no emission for first record")
	}

	second := lineRecord{timestamp: later}
	emit := buffer.push(second)
	if emit == nil {
		t.Fatalf("expected emission after second record")
	}
	if emit.delta != 2*time.Second {
		t.Fatalf("unexpected delta: %v", emit.delta)
	}
	if emit.elapsed != 0 {
		t.Fatalf("unexpected elapsed for first line: %v", emit.elapsed)
	}
	if emit.line != 1 {
		t.Fatalf("unexpected line number: %d", emit.line)
	}

	flush := buffer.flush()
	if flush == nil {
		t.Fatalf("expected flush emission for final record")
	}
	if flush.delta != 0 {
		t.Fatalf("unexpected final delta: %v", flush.delta)
	}
	if flush.elapsed != 2*time.Second {
		t.Fatalf("unexpected final elapsed: %v", flush.elapsed)
	}
	if flush.line != 2 {
		t.Fatalf("unexpected final line number: %d", flush.line)
	}

	if buffer.flush() != nil {
		t.Fatalf("expected no emission after final flush")
	}
}

func TestTextEmitterEmit(t *testing.T) {
	tpl, err := template.Parse("{elapsed:.1f}s {delta:.1f}s #{line} {}")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	var buf bytes.Buffer
	emitter := newTextEmitter(tpl, &buf)

	rec := lineRecord{text: "hello", hasNewline: true, timestamp: time.Unix(0, 0)}
	em := emission{record: rec, delta: 1500 * time.Millisecond, elapsed: 2500 * time.Millisecond, line: 3}
	if err := emitter.emit(em); err != nil {
		t.Fatalf("emit returned error: %v", err)
	}

	want := "2.5s 1.5s #3 hello\n"
	if buf.String() != want {
		t.Fatalf("unexpected output: got %q want %q", buf.String(), want)
	}

	buf.Reset()
	rec.hasNewline = false
	em.record = rec
	if err := emitter.emit(em); err != nil {
		t.Fatalf("emit returned error: %v", err)
	}
	if buf.String() != "2.5s 1.5s #3 hello" {
		t.Fatalf("unexpected output without newline: %q", buf.String())
	}
}
