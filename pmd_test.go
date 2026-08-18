package main

import (
	"encoding/json"
	"testing"
)

const sampleSARIF = `{
  "version": "2.1.0",
  "$schema": "http://json.schemastore.org/sarif-2.1.0",
  "runs": [
    {
      "tool": {"driver": {"name": "eslint"}},
      "results": [
        {
          "ruleId": "matched",
          "level": "warning",
          "message": {"text": "changed line"},
          "locations": [
            {"physicalLocation": {"artifactLocation": {"uri": "force-app/main/default/foo.js"}, "region": {"startLine": 10}}}
          ]
        },
        {
          "ruleId": "unmatched",
          "level": "warning",
          "message": {"text": "untouched line"},
          "locations": [
            {"physicalLocation": {"artifactLocation": {"uri": "force-app/main/default/foo.js"}, "region": {"startLine": 99}}}
          ]
        }
      ]
    }
  ]
}`

func TestFilterKeepsMatchingResults(t *testing.T) {
	var report pmdReport
	if err := json.Unmarshal([]byte(sampleSARIF), &report); err != nil {
		t.Fatalf("unmarshal sample: %v", err)
	}

	checker := &diffChecker{changes: map[string][]int{
		"force-app/main/default/foo.js": {10},
	}}

	got := report.filter(checker)

	if len(got.Runs) != 1 {
		t.Fatalf("runs = %d, want 1", len(got.Runs))
	}
	if len(got.Runs[0].Results) != 1 {
		t.Fatalf("results = %d, want 1", len(got.Runs[0].Results))
	}

	var kept sarifResult
	if err := json.Unmarshal(got.Runs[0].Results[0], &kept); err != nil {
		t.Fatalf("unmarshal kept result: %v", err)
	}
	if kept.RuleId != "matched" {
		t.Errorf("kept ruleId = %q, want matched", kept.RuleId)
	}

	if got.Version != "2.1.0" {
		t.Errorf("version = %q, want 2.1.0", got.Version)
	}
	if got.Schema != "http://json.schemastore.org/sarif-2.1.0" {
		t.Errorf("schema = %q, not preserved", got.Schema)
	}
	if len(got.Runs[0].Tool) == 0 {
		t.Error("tool metadata not preserved")
	}
}

func TestFilterNoMatches(t *testing.T) {
	var report pmdReport
	if err := json.Unmarshal([]byte(sampleSARIF), &report); err != nil {
		t.Fatalf("unmarshal sample: %v", err)
	}

	checker := &diffChecker{changes: map[string][]int{}}

	got := report.filter(checker)

	if len(got.Runs) != 1 {
		t.Fatalf("runs = %d, want 1", len(got.Runs))
	}
	if len(got.Runs[0].Results) != 0 {
		t.Fatalf("results = %d, want 0", len(got.Runs[0].Results))
	}

	data, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal filtered report: %v", err)
	}
	if !json.Valid(data) {
		t.Error("filtered report is not valid json")
	}
}
