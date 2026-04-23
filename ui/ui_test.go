package ui

import (
	"errors"
	"os"
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

func TestCanPrompt_NonInteractiveEnvVar(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     bool
	}{
		{"not set", "", true},            // When not set, depends on terminal (true in test env with TTY)
		{"set to 1", "1", false},         // Explicitly non-interactive
		{"set to true", "true", false},   // Explicitly non-interactive
		{"set to 0", "0", true},          // 0 should not disable prompts
		{"set to false", "false", true},  // false should not disable prompts
		{"set to other", "yes", true},    // Other values should not disable prompts
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore the environment variable
			oldValue, wasSet := os.LookupEnv("STEP_NON_INTERACTIVE")
			defer func() {
				if wasSet {
					os.Setenv("STEP_NON_INTERACTIVE", oldValue)
				} else {
					os.Unsetenv("STEP_NON_INTERACTIVE")
				}
			}()

			if tt.envValue == "" {
				os.Unsetenv("STEP_NON_INTERACTIVE")
			} else {
				os.Setenv("STEP_NON_INTERACTIVE", tt.envValue)
			}

			got := CanPrompt()

			// When STEP_NON_INTERACTIVE is "1" or "true", CanPrompt should always return false
			// For other values, it depends on whether a terminal is available
			if tt.envValue == "1" || tt.envValue == "true" {
				if got != false {
					t.Errorf("CanPrompt() = %v, want %v when STEP_NON_INTERACTIVE=%q", got, false, tt.envValue)
				}
			}
		})
	}
}
