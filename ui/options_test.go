package ui

import "testing"

func TestWithMinLength(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "negative",
			length: -5,
		},
		{
			name:   "zero",
			length: 0,
		},
		{
			name:   "positive",
			length: 11,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &options{}
			WithMinLength(tt.length)(o)
			if o.minLength != tt.length {
				t.Errorf("want %v, but got %v", tt.length, o.minLength)
			}
		})
	}
}
