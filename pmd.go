package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type pmdReport struct {
	Version string   `json:"version"`
	Schema  string   `json:"$schema"`
	Runs    []pmdRun `json:"runs"`
}

type pmdRun struct {
	Tool        json.RawMessage   `json:"tool"`
	Results     []json.RawMessage `json:"results"`
	Invocations json.RawMessage   `json:"invocations,omitempty"`
}

type sarifResult struct {
	Level   string `json:"level"`
	RuleId  string `json:"ruleId"`
	Message struct {
		Text string `json:"text"`
	} `json:"message"`
	Locations []struct {
		PhysicalLocation struct {
			ArtifactLocation struct {
				Uri string `json:"uri"`
			} `json:"artifactLocation"`
			Region struct {
				StartLine int `json:"startLine"`
				EndLine   int `json:"endLine"`
			} `json:"region"`
		} `json:"physicalLocation"`
	} `json:"locations"`
}

func readPMDFile(path string) (*pmdReport, error) {
	var report pmdReport

	f, err := os.Open(path)
	if err != nil {
		return &report, fmt.Errorf("could not open pmd report: %w", err)
	}
	defer f.Close()

	if err = json.NewDecoder(f).Decode(&report); err != nil {
		return &report, fmt.Errorf("could not parse pmd report: %w", err)
	}

	return &report, err
}

func (r *pmdReport) violations() ([]violation, error) {
	var result []violation

	for _, run := range r.Runs {
		for _, raw := range run.Results {
			var res sarifResult
			if err := json.Unmarshal(raw, &res); err != nil {
				return nil, fmt.Errorf("could not parse sarif result: %w", err)
			}
			for _, l := range res.Locations {
				result = append(result, violation{
					rule:     res.RuleId,
					line:     l.PhysicalLocation.Region.StartLine,
					filename: stripPath(l.PhysicalLocation.ArtifactLocation.Uri, "force-app"),
					message:  res.Message.Text,
					severity: res.Level,
				})
			}
		}
	}

	return result, nil
}

func (r *pmdReport) filter(c contentChecker) *pmdReport {
	out := &pmdReport{
		Version: r.Version,
		Schema:  r.Schema,
	}

	for _, run := range r.Runs {
		filtered := pmdRun{
			Tool:        run.Tool,
			Invocations: run.Invocations,
			Results:     []json.RawMessage{},
		}
		for _, raw := range run.Results {
			if resultMatches(raw, c) {
				filtered.Results = append(filtered.Results, raw)
			}
		}
		out.Runs = append(out.Runs, filtered)
	}

	return out
}

func resultMatches(raw json.RawMessage, c contentChecker) bool {
	var res sarifResult
	if err := json.Unmarshal(raw, &res); err != nil {
		return false
	}

	for _, l := range res.Locations {
		filename := stripPath(l.PhysicalLocation.ArtifactLocation.Uri, "force-app")
		if c.contains(filename, l.PhysicalLocation.Region.StartLine) {
			return true
		}
	}

	return false
}

func writeSARIFFile(path string, r *pmdReport) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("could not encode sarif: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("could not write sarif file: %w", err)
	}

	return nil
}

func stripPath(input, keyword string) string {
	index := strings.Index(input, keyword)
	if index != -1 {
		return input[index:]
	}
	return input
}
