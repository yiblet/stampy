package template

import (
	"fmt"
	"testing"
	"time"
)

func TestTemplateRenderVariants(t *testing.T) {
	base := time.Date(2024, 7, 4, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name  string
		tpl   string
		state StampState
		want  string
	}{
		{
			name: "elapsed with placeholder",
			tpl:  "{elapsed:.1f}s {}",
			state: StampState{
				Now:      base,
				Elapsed:  1250 * time.Millisecond,
				LineText: "build start",
			},
			want: "1.2s build start",
		},
		{
			name: "delta default precision",
			tpl:  "Δ{delta}s {}",
			state: StampState{
				Now:      base,
				Delta:    2200 * time.Millisecond,
				LineText: "compile",
			},
			want: "Δ2.2s compile",
		},
		{
			name: "time go layout",
			tpl:  "{time:2006-01-02 15:04} {}",
			state: StampState{
				Now:      base,
				LineText: "snapshot",
			},
			want: "2024-07-04 12:00 snapshot",
		},
		{
			name: "time named layout",
			tpl:  "{time:iso} {}",
			state: StampState{
				Now:      base,
				LineText: "tick",
			},
			want: "2024-07-04T12:00:00Z tick",
		},
		{
			name: "iso token",
			tpl:  "{iso} {}",
			state: StampState{
				Now:      base,
				LineText: "tick",
			},
			want: "2024-07-04T12:00:00Z tick",
		},
		{
			name: "unix token default",
			tpl:  "{unix} {}",
			state: StampState{
				Now:      base,
				LineText: "tick",
			},
			want: fmt.Sprintf("%d tick", base.Unix()),
		},
		{
			name: "unix token formatted",
			tpl:  "{unix:.2f} {}",
			state: StampState{
				Now:      base,
				LineText: "tick",
			},
			want: fmt.Sprintf("%.2f tick", float64(base.Unix())),
		},
		{
			name: "time unix",
			tpl:  "{time:unix} {}",
			state: StampState{
				Now:      base,
				LineText: "tick",
			},
			want: fmt.Sprintf("%d tick", base.Unix()),
		},
		{
			name: "time date layout",
			tpl:  "{time:%Y-%m-%d %H:%M:%S} {}",
			state: StampState{
				Now:      base,
				LineText: "tick",
			},
			want: "2024-07-04 12:00:00 tick",
		},
		{
			name: "line number",
			tpl:  "#{line} {elapsed:.0f}s {}",
			state: StampState{
				Now:      base,
				Elapsed:  3 * time.Second,
				Line:     5,
				LineText: "result",
			},
			want: "#5 3s result",
		},
		{
			name: "implicit line append",
			tpl:  "{elapsed:.1f}s",
			state: StampState{
				Now:      base,
				Elapsed:  1700 * time.Millisecond,
				LineText: "done",
			},
			want: "1.7s done",
		},
		{
			name: "escaped braces",
			tpl:  "{{ {elapsed:.0f} }} {}",
			state: StampState{
				Now:      base,
				Elapsed:  5 * time.Second,
				LineText: "value",
			},
			want: "{ 5 } value",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tpl, err := Parse(tc.tpl)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}
			got := tpl.Render(tc.state)
			if got != tc.want {
				t.Fatalf("render mismatch: got %q want %q", got, tc.want)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	cases := []struct {
		name string
		tpl  string
	}{
		{name: "unterminated", tpl: "{elapsed"},
		{name: "unknown token", tpl: "{unknown} {}"},
		{name: "duplicate line placeholder", tpl: "{} {}"},
		{name: "bad duration modifier", tpl: "{elapsed:{}}"},
		{name: "unsupported date directive", tpl: "{time:%Q}"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Parse(tc.tpl); err == nil {
				t.Fatalf("expected error for template %q", tc.tpl)
			}
		})
	}
}

func TestConvertDateLayout(t *testing.T) {
	t.Run("basic datetime", func(t *testing.T) {
		got, err := convertDateLayout("%Y-%m-%d %H:%M:%S")
		if err != nil {
			t.Fatalf("convert failed: %v", err)
		}
		want := "2006-01-02 15:04:05"
		if got != want {
			t.Fatalf("unexpected conversion: got %q want %q", got, want)
		}
	})

	t.Run("escaped percent", func(t *testing.T) {
		got, err := convertDateLayout("%%Y")
		if err != nil {
			t.Fatalf("convert failed: %v", err)
		}
		if got != "%Y" {
			t.Fatalf("unexpected conversion: got %q want %q", got, "%Y")
		}
	})

	t.Run("microseconds", func(t *testing.T) {
		got, err := convertDateLayout("%Y-%m-%dT%H:%M:%S.%fZ")
		if err != nil {
			t.Fatalf("convert failed: %v", err)
		}
		want := "2006-01-02T15:04:05.000000Z"
		if got != want {
			t.Fatalf("unexpected conversion: got %q want %q", got, want)
		}
	})

	t.Run("microseconds without dot", func(t *testing.T) {
		got, err := convertDateLayout("%Y-%m-%dT%H:%M:%S%fZ")
		if err != nil {
			t.Fatalf("convert failed: %v", err)
		}
		want := "2006-01-02T15:04:05.000000Z"
		if got != want {
			t.Fatalf("unexpected conversion: got %q want %q", got, want)
		}
	})

	t.Run("unsupported directive", func(t *testing.T) {
		if _, err := convertDateLayout("%Q"); err == nil {
			t.Fatalf("expected error for unsupported directive")
		}
	})
}
