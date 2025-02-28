package ui

import (
	"errors"
	"testing"
)

func Test_promptRun(t *testing.T) {
	promptRunner := func(input []string, err error) func() (string, error) {
		i := 0
		return func() (string, error) {
			ret := input[i]
			i++
			return ret, err
		}
	}

	tests := []struct {
		name      string
		minLength int
		promptRun func() (string, error)
		want      string
		wantErr   bool
	}{
		{
			name:      "prompt-error",
			minLength: -5,
			promptRun: promptRunner([]string{"foobar"}, errors.New("prompt-error")),
			want:      "foobar",
			wantErr:   true,
		},
		{
			name:      "negative",
			minLength: -5,
			promptRun: promptRunner([]string{"foobar"}, nil),
			want:      "foobar",
			wantErr:   false,
		},
		{
			name:      "zero",
			minLength: 0,
			promptRun: promptRunner([]string{"foobar"}, nil),
			want:      "foobar",
			wantErr:   false,
		},
		{
			name:      "greater-than-min-length",
			minLength: 5,
			promptRun: promptRunner([]string{"foobar"}, nil),
			want:      "foobar",
			wantErr:   false,
		},
		{
			name:      "equal-min-length",
			minLength: 6,
			promptRun: promptRunner([]string{"foobar"}, nil),
			want:      "foobar",
			wantErr:   false,
		},
		{
			name:      "less-than-min-length",
			minLength: 8,
			promptRun: promptRunner([]string{"pass", "foobar", "password"}, nil),
			want:      "password",
			wantErr:   false,
		},
		{
			name:      "ignore-post-whitespace-characters",
			minLength: 7,
			promptRun: promptRunner([]string{"pass   ", "foobar ", "password  "}, nil),
			want:      "password",
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := runPrompt(tt.promptRun, &options{minLength: tt.minLength})
			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Errorf("expected error=%v, but got error=%v", tt.wantErr, err)
				return
			}
			if gotErr {
				return
			}
			if val != tt.want {
				t.Errorf("expected %v, but got %v", tt.want, val)
			}
		})
	}
}
