package template

import (
	"fmt"
	"strings"
)

var dateMappings = map[byte]string{
	'Y': "2006",
	'y': "06",
	'm': "01",
	'd': "02",
	'H': "15",
	'I': "03",
	'M': "04",
	'S': "05",
	'z': "-0700",
	'Z': "MST",
	'j': "002",
	'b': "Jan",
	'B': "January",
	'a': "Mon",
	'A': "Monday",
	'p': "PM",
}

func convertDateLayout(layout string) (string, error) {
	var b strings.Builder
	var last byte
	for i := 0; i < len(layout); i++ {
		ch := layout[i]
		if ch != '%' {
			b.WriteByte(ch)
			last = ch
			continue
		}
		i++
		if i >= len(layout) {
			return "", fmt.Errorf("incomplete date directive at end of layout")
		}
		directive := layout[i]
		if directive == '%' {
			b.WriteByte('%')
			last = '%'
			continue
		}
		if directive == 'f' {
			// `%f` from date(1) represents microseconds; Go layouts use an explicit fractional
			// second pattern so we ensure exactly one '.' precedes six zeros.
			if last != '.' {
				b.WriteByte('.')
			}
			b.WriteString("000000")
			last = '0'
			continue
		}
		goLayout, ok := dateMappings[directive]
		if !ok {
			return "", fmt.Errorf("unsupported date directive '%%%c'", directive)
		}
		b.WriteString(goLayout)
		if len(goLayout) > 0 {
			last = goLayout[len(goLayout)-1]
		}
	}
	return b.String(), nil
}
