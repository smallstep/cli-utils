package ui

import (
	"testing"
)

func TestDNS(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "localhost",
			input:   "localhost",
			wantErr: false,
		},
		{
			name:    "example.com",
			input:   "example.com",
			wantErr: false,
		},
		{
			name:    "ca.smallstep.com",
			input:   "ca.smallstep.com",
			wantErr: false,
		},
		{
			name:    "localhost-with-port",
			input:   "localhost:443",
			wantErr: true,
		},
		{
			name:    "quad1",
			input:   "1.1.1.1",
			wantErr: false,
		},
		{
			name:    "ipv4-localhost",
			input:   "127.0.0.1",
			wantErr: false,
		},
		{
			name:    "ipv6-localhost",
			input:   "::1",
			wantErr: false,
		},
		{
			name:    "ipv6-localhost-brackets",
			input:   "[::1]",
			wantErr: false,
		},
		{
			name:    "ipv6",
			input:   "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			wantErr: false,
		},
		{
			name:    "ipv6-brackets",
			input:   "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]",
			wantErr: false,
		},
		{
			name:    "ipv6-shortened",
			input:   "2001:0db8:85a3::8a2e:0370:7334",
			wantErr: false,
		},
		{
			name:    "ipv6-shortened-brackets",
			input:   "[2001:0db8:85a3::8a2e:0370:7334]",
			wantErr: false,
		},
		{
			name:    "ipv6-shortened-brackets-missing-end",
			input:   "[2001:0db8:85a3::8a2e:0370:7334",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := DNS()(tt.input) != nil
			if gotErr != tt.wantErr {
				t.Errorf("DNS()(%s) = %v, want %v", tt.input, gotErr, tt.wantErr)
			}
		})
	}
}

func TestMinLen(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		input   string
		wantErr bool
	}{
		{
			name:    "negative",
			length:  -5,
			input:   "foobar",
			wantErr: false,
		},
		{
			name:    "zero",
			length:  0,
			input:   "localhost",
			wantErr: false,
		},
		{
			name:    "greater-than-min-length",
			length:  5,
			input:   "foobar",
			wantErr: false,
		},
		{
			name:    "equal-min-length",
			length:  6,
			input:   "foobar",
			wantErr: false,
		},
		{
			name:    "less-than-min-length",
			length:  8,
			input:   "foobar",
			wantErr: true,
		},
		{
			name:    "ignore-whitespace-characters",
			length:  15,
			input:   "  p \t\n#$%@#$ a s s  ",
			wantErr: true,
		},
		{
			name:    "ignore-whitespace-characters-ok",
			length:  10,
			input:   "  p \t\n#$%@#$ a s s  ",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := MinLen(tt.length)(tt.input) != nil
			if gotErr != tt.wantErr {
				t.Errorf("MinLen(%v)(%s) = %v, want %v", tt.length, tt.input, gotErr, tt.wantErr)
			}
		})
	}
}
