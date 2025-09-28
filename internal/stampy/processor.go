package stampy

import (
	"io"
	"time"

	"github.com/yiblet/stampy/internal/stampy/template"
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
