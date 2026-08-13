package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantExitCode int
		wantContains []string
	}{
		{
			name:         "no args prints usage and exits 2",
			args:         nil,
			wantExitCode: 2,
			wantContains: []string{"ghp <command>"},
		},
		{
			name:         "dev starts the development server",
			args:         []string{"dev"},
			wantExitCode: 0,
			wantContains: []string{"Starting development server..."},
		},
		{
			name:         "build builds the project",
			args:         []string{"build"},
			wantExitCode: 0,
			wantContains: []string{"Building GHP project...", "GHP project built"},
		},
		{
			name:         "help prints usage",
			args:         []string{"help"},
			wantExitCode: 0,
			wantContains: []string{"ghp <command>"},
		},
		{
			name:         "unknown command prints usage and exits 2",
			args:         []string{"bogus"},
			wantExitCode: 2,
			wantContains: []string{"Command unknown", "ghp <command>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer

			gotExitCode := run(tt.args, &out)

			if gotExitCode != tt.wantExitCode {
				t.Errorf("run(%v) exit code = %d, want %d", tt.args, gotExitCode, tt.wantExitCode)
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(out.String(), want) {
					t.Errorf("run(%v) output missing %q\ngot: %s", tt.args, want, out.String())
				}
			}
		})
	}
}

func TestPrintUsage(t *testing.T) {
	var out bytes.Buffer

	printUsage(&out)

	for _, want := range []string{"ghp <command>", "dev", "build", "help"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("printUsage() output missing %q\ngot: %s", want, out.String())
		}
	}
}
