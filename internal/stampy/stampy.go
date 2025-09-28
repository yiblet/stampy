package stampy

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/yiblet/stampy/internal/stampy/template"
)

const defaultTemplate = "{iso}: {}"

// Options captures the configuration used when running the timestamping workflow.
type Options struct {
	Template         string
	TemplateProvided bool
	Input            string
	Output           string
	JSONKey          string
}

// Run executes the timestamping workflow using the system clock.
func Run(opts Options) error {
	return RunWithClock(opts, time.Now)
}

// RunWithClock executes the timestamping workflow with a provided clock, making it testable.
func RunWithClock(opts Options, nowFn func() time.Time) (err error) {
	tplString := opts.Template
	if !opts.TemplateProvided || tplString == "" {
		tplString = defaultTemplate
	}

	tpl, err := template.Parse(tplString)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	reader, writer, cleanup, err := createIO(opts.Input, opts.Output)
	if err != nil {
		return err
	}

	defer func() {
		if cerr := cleanup(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	return processLines(reader, writer, tpl, opts, nowFn)
}

// createIO wires up the appropriate reader and writer based on the provided
// paths and returns a cleanup function that closes any opened files.
func createIO(in string, out string) (io.Reader, io.Writer, func() error, error) {
	var inFile = os.Stdin
	var outFile = os.Stdout

	mustClose := [](func() error){}
	if in != "" {
		f, err := os.Open(in)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to open input file: %v", err)
		}
		inFile = f
		mustClose = append(mustClose, inFile.Close)
	}

	if out != "" {
		f, err := os.Create(out)
		if err != nil {
			for _, close := range mustClose {
				if err := close(); err != nil {
					return nil, nil, nil, err
				}
			}
			return nil, nil, nil, fmt.Errorf("failed to open output file: %v", err)
		}
		outFile = f
		mustClose = append(mustClose, outFile.Close)
	}
	return inFile, outFile, func() error {
		errs := []error{}
		for _, close := range mustClose {
			if err := close(); err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	}, nil
}

type lineRecord struct {
	text       string
	hasNewline bool
	timestamp  time.Time
}

func processLines(reader io.Reader, writer io.Writer, tpl template.Template, opts Options, nowFn func() time.Time) error {
	bufreader := bufio.NewReader(reader)
	buffer := newLineBuffer()

	// Select emitter based on whether JSONL mode is enabled
	var emitter interface {
		emit(emission) error
	}
	if opts.JSONKey != "" {
		emitter = newJSONEmitter(tpl, writer, opts.JSONKey)
	} else {
		emitter = newTextEmitter(tpl, writer)
	}

	for {
		line, err := bufreader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("read line: %w", err)
		}

		if len(line) == 0 && errors.Is(err, io.EOF) {
			break
		}

		record := lineRecord{
			text:       strings.TrimSuffix(line, "\n"),
			hasNewline: strings.HasSuffix(line, "\n"),
			timestamp:  nowFn(),
		}

		if emit := buffer.push(record); emit != nil {
			if err := emitter.emit(*emit); err != nil {
				return err
			}
		}

		if errors.Is(err, io.EOF) {
			break
		}
	}

	if emit := buffer.flush(); emit != nil {
		if err := emitter.emit(*emit); err != nil {
			return err
		}
	}

	return nil
}
