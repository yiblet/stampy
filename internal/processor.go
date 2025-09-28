package internal

import (
	"encoding/json"
	"io"
	"time"

	"github.com/yiblet/stampy/internal/template"
)

type emission struct {
	record  lineRecord
	delta   time.Duration
	elapsed time.Duration
	line    int
}

type lineBuffer struct {
	start      time.Time
	haveStart  bool
	pending    lineRecord
	hasPending bool
	lineNumber int
}

func newLineBuffer() *lineBuffer {
	return &lineBuffer{}
}

func (b *lineBuffer) push(record lineRecord) *emission {
	if !b.haveStart {
		b.start = record.timestamp
		b.haveStart = true
	}

	if !b.hasPending {
		b.pending = record
		b.hasPending = true
		return nil
	}

	emit := b.prepareEmission(b.pending, record.timestamp.Sub(b.pending.timestamp))
	b.pending = record
	return emit
}

func (b *lineBuffer) flush() *emission {
	if !b.hasPending {
		return nil
	}
	emit := b.prepareEmission(b.pending, 0)
	b.hasPending = false
	return emit
}

func (b *lineBuffer) prepareEmission(record lineRecord, delta time.Duration) *emission {
	b.lineNumber++
	elapsed := record.timestamp.Sub(b.start)
	return &emission{
		record:  record,
		delta:   delta,
		elapsed: elapsed,
		line:    b.lineNumber,
	}
}

type textEmitter struct {
	tpl    template.Template
	writer io.Writer
}

func newTextEmitter(tpl template.Template, writer io.Writer) textEmitter {
	return textEmitter{tpl: tpl, writer: writer}
}

func (e textEmitter) emit(em emission) error {
	rendered := e.tpl.Render(template.StampState{
		Now:      em.record.timestamp,
		Delta:    em.delta,
		Elapsed:  em.elapsed,
		Line:     em.line,
		LineText: em.record.text,
	})
	if _, err := io.WriteString(e.writer, rendered); err != nil {
		return err
	}
	if em.record.hasNewline {
		if _, err := io.WriteString(e.writer, "\n"); err != nil {
			return err
		}
	}
	return nil
}

// jsonEmitter outputs JSONL format by stamping and merging/wrapping JSON objects.
type jsonEmitter struct {
	tpl     template.Template
	writer  io.Writer
	jsonKey string
}

// newJSONEmitter creates a new JSONL emitter with the given template, writer, and JSON key.
func newJSONEmitter(tpl template.Template, writer io.Writer, jsonKey string) jsonEmitter {
	return jsonEmitter{tpl: tpl, writer: writer, jsonKey: jsonKey}
}

// emit processes the emission by rendering the template stamp and merging/wrapping with JSON.
func (e jsonEmitter) emit(em emission) error {
	// Render the template to get the stamp string
	// In JSONL mode, we don't want the line text auto-appended to the template
	stamp := e.tpl.Render(template.StampState{
		Now:      em.record.timestamp,
		Delta:    em.delta,
		Elapsed:  em.elapsed,
		Line:     em.line,
		LineText: "", // Empty to prevent auto-appending line text
	})

	// Process the input line as JSON
	result, err := e.processJSONLine(em.record.text, stamp)
	if err != nil {
		return err
	}

	// Write the result as compact JSON
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return err
	}

	if _, err := e.writer.Write(jsonBytes); err != nil {
		return err
	}

	// Only add newline if the original record had one
	if em.record.hasNewline {
		if _, err := e.writer.Write([]byte("\n")); err != nil {
			return err
		}
	}

	return nil
}

// processJSONLine handles the JSON parsing and merging/wrapping logic.
func (e jsonEmitter) processJSONLine(line, stamp string) (any, error) {
	// Try to parse the line as JSON
	var parsed any
	if err := json.Unmarshal([]byte(line), &parsed); err != nil {
		// Parse failed: wrap as {"jsonKey": stamp, "line": originalString}
		return map[string]any{
			e.jsonKey: stamp,
			"line":    line,
		}, nil
	}

	// Parse succeeded: handle based on type
	switch v := parsed.(type) {
	case map[string]any:
		// JSON object: set jsonObj[jsonKey] = stamp
		v[e.jsonKey] = stamp
		return v, nil
	default:
		// Primitive or array: wrap as {"jsonKey": stamp, "line": value}
		return map[string]any{
			e.jsonKey: stamp,
			"line":    v,
		}, nil
	}
}
