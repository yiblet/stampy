package template

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// StampState carries the data needed to render a template instance.
type StampState struct {
	Now      time.Time
	Delta    time.Duration
	Elapsed  time.Duration
	Line     int
	LineText string
}

// Template renders brace-based stamp expressions.
type Template struct {
	segments           []segment
	hasLinePlaceholder bool
}

// Parse builds a Template from the provided brace expression.
func Parse(input string) (Template, error) {
	p := parser{input: input}
	tpl, err := p.parse()
	if err != nil {
		return Template{}, err
	}
	return tpl, nil
}

// Render evaluates the template using the supplied state.
func (t Template) Render(state StampState) string {
	var b strings.Builder
	for _, seg := range t.segments {
		seg.append(&b, state)
	}

	if !t.hasLinePlaceholder && state.LineText != "" {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(state.LineText)
	}

	return b.String()
}

type segment interface {
	append(*strings.Builder, StampState)
}

type literalSegment struct {
	value string
}

func (l literalSegment) append(b *strings.Builder, _ StampState) {
	b.WriteString(l.value)
}

type lineSegment struct{}

func (lineSegment) append(b *strings.Builder, state StampState) {
	b.WriteString(state.LineText)
}

type tokenSegment struct {
	eval tokenEvaluator
}

func (t tokenSegment) append(b *strings.Builder, state StampState) {
	b.WriteString(t.eval(state))
}

type tokenEvaluator func(StampState) string

type parser struct {
	input string
	pos   int
}

func (p *parser) parse() (Template, error) {
	segments := []segment{}
	var literal strings.Builder
	hasLine := false

	flushLiteral := func() {
		if literal.Len() == 0 {
			return
		}
		segments = append(segments, literalSegment{value: literal.String()})
		literal.Reset()
	}

	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		switch ch {
		case '{':
			if p.pos+1 < len(p.input) && p.input[p.pos+1] == '{' {
				literal.WriteByte('{')
				p.pos += 2
				continue
			}
			flushLiteral()
			tokenContent, err := p.consumeToken()
			if err != nil {
				return Template{}, err
			}
			if tokenContent == "" {
				if hasLine {
					return Template{}, fmt.Errorf("template contains multiple '{}' placeholders")
				}
				hasLine = true
				segments = append(segments, lineSegment{})
				continue
			}
			tokenSeg, err := buildTokenSegment(tokenContent)
			if err != nil {
				return Template{}, err
			}
			segments = append(segments, tokenSeg)
		case '}':
			if p.pos+1 < len(p.input) && p.input[p.pos+1] == '}' {
				literal.WriteByte('}')
				p.pos += 2
				continue
			}
			literal.WriteByte('}')
			p.pos++
		default:
			literal.WriteByte(ch)
			p.pos++
		}
	}

	flushLiteral()

	return Template{segments: segments, hasLinePlaceholder: hasLine}, nil
}

func (p *parser) consumeToken() (string, error) {
	start := p.pos + 1
	depth := 1
	i := start
	for ; i < len(p.input); i++ {
		if p.input[i] == '{' {
			depth++
			continue
		}
		if p.input[i] == '}' {
			depth--
			if depth == 0 {
				token := p.input[start:i]
				p.pos = i + 1
				return token, nil
			}
		}
	}
	return "", fmt.Errorf("unterminated '{' in template")
}

func buildTokenSegment(raw string) (segment, error) {
	name := raw
	arg := ""
	if idx := strings.IndexRune(raw, ':'); idx != -1 {
		name = raw[:idx]
		arg = raw[idx+1:]
	}
	name = strings.TrimSpace(name)
	arg = strings.TrimSpace(arg)

	if name == "" {
		return nil, fmt.Errorf("empty token in template")
	}

	switch name {
	case "elapsed":
		evaluator, err := durationEvaluator(arg, func(state StampState) time.Duration { return state.Elapsed })
		if err != nil {
			return nil, err
		}
		return tokenSegment{eval: evaluator}, nil
	case "delta":
		evaluator, err := durationEvaluator(arg, func(state StampState) time.Duration { return state.Delta })
		if err != nil {
			return nil, err
		}
		return tokenSegment{eval: evaluator}, nil
	case "time":
		layout, unixStamp, err := resolveTimeLayout(arg)
		if err != nil {
			return nil, err
		}
		if unixStamp {
			evaluator, err := unixEvaluator("")
			if err != nil {
				return nil, err
			}
			return tokenSegment{eval: evaluator}, nil
		}
		return tokenSegment{eval: func(state StampState) string {
			return state.Now.Format(layout)
		}}, nil
	case "iso":
		layout, _, err := resolveTimeLayout("iso")
		if err != nil {
			return nil, err
		}
		return tokenSegment{eval: func(state StampState) string {
			return state.Now.Format(layout)
		}}, nil
	case "unix":
		evaluator, err := unixEvaluator(arg)
		if err != nil {
			return nil, err
		}
		return tokenSegment{eval: evaluator}, nil
	case "line":
		return tokenSegment{eval: func(state StampState) string {
			return strconv.Itoa(state.Line)
		}}, nil
	default:
		return nil, fmt.Errorf("unknown token '%s'", name)
	}
}

func durationEvaluator(modifier string, getter func(StampState) time.Duration) (tokenEvaluator, error) {
	fmtStr := "%.1f"
	if modifier != "" {
		if strings.ContainsAny(modifier, "{}") {
			return nil, fmt.Errorf("invalid duration modifier '%s'", modifier)
		}
		fmtStr = "%" + modifier
	} else {
		fmtStr = "%.1f"
	}

	return func(state StampState) string {
		seconds := getter(state).Seconds()
		return fmt.Sprintf(fmtStr, seconds)
	}, nil
}

func normalizeLayout(layout string) (string, error) {
	if strings.Contains(layout, "%") {
		return convertDateLayout(layout)
	}
	return layout, nil
}

func resolveTimeLayout(arg string) (layout string, unix bool, err error) {
	if arg == "" {
		return time.RFC3339, false, nil
	}
	switch strings.ToLower(arg) {
	case "iso", "iso8601":
		return time.RFC3339, false, nil
	case "iso8601nano", "isonano":
		return time.RFC3339Nano, false, nil
	case "unix", "unixs":
		return "", true, nil
	}

	normalized, err := normalizeLayout(arg)
	if err != nil {
		return "", false, err
	}
	return normalized, false, nil
}

func unixEvaluator(modifier string) (tokenEvaluator, error) {
	if modifier == "" {
		return func(state StampState) string {
			return strconv.FormatInt(state.Now.Unix(), 10)
		}, nil
	}
	if strings.ContainsAny(modifier, "{}") {
		return nil, fmt.Errorf("invalid unix modifier '%s'", modifier)
	}
	fmtStr := "%" + modifier
	return func(state StampState) string {
		seconds := float64(state.Now.UnixNano()) / float64(time.Second)
		// Avoid printing negative zero when time is before epoch.
		if seconds == 0 {
			seconds = 0
		}
		return fmt.Sprintf(fmtStr, seconds)
	}, nil
}
