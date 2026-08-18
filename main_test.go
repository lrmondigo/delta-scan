package main

import "testing"

func TestParseOptions(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"report required", []string{"-d", "change.diff"}, true},
		{"valid minimal", []string{"-r", "report.sarif"}, false},
		{"export requires file", []string{"-r", "report.sarif", "-x"}, true},
		{"file requires export", []string{"-r", "report.sarif", "-f", "out.sarif"}, true},
		{"valid export", []string{"-r", "report.sarif", "-x", "-f", "out.sarif"}, false},
		{"unexpected positional", []string{"-r", "report.sarif", "extra.json"}, true},
		{"help skips validation", []string{"-h"}, false},
		{"unknown flag", []string{"-z"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseOptions(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseOptions(%v) err = %v, wantErr %v", tt.args, err, tt.wantErr)
			}
		})
	}
}
