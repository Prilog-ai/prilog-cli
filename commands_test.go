package main

import (
	"strings"
	"testing"
)

func TestNormalizeListStatus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		want  string
	}{
		{input: "", want: ""},
		{input: "all", want: ""},
		{input: " Pending ", want: "pending"},
		{input: "processing", want: "processing"},
		{input: "completed", want: "completed"},
		{input: "failed", want: "failed"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got, err := normalizeListStatus(tc.input)
			if err != nil {
				t.Fatalf("normalizeListStatus returned error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("normalizeListStatus(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNormalizeListStatusRejectsUnsupportedFilter(t *testing.T) {
	t.Parallel()

	_, err := normalizeListStatus("open")
	if err == nil || !strings.Contains(err.Error(), "unsupported list filter") {
		t.Fatalf("unsupported filter error = %v", err)
	}
}

func TestConfigRejectsPublicArgumentsWithoutExposingInternalOverrides(t *testing.T) {
	t.Parallel()

	c := newCLI(defaultAPIURL, ".", strings.NewReader(""), nil, nil)
	err := c.config([]string{"bad"})
	if err == nil {
		t.Fatal("config returned nil error")
	}
	if strings.Contains(err.Error(), "api-url") {
		t.Fatalf("config error exposed internal override: %v", err)
	}
}

func TestReadIngestPayloadFromStdin(t *testing.T) {
	t.Parallel()

	body, filename, err := readIngestPayload(nil, strings.NewReader("ERROR line"))
	if err != nil {
		t.Fatalf("readIngestPayload returned error: %v", err)
	}
	if string(body) != "ERROR line" || filename != "stdin" {
		t.Fatalf("payload = %q filename=%q", string(body), filename)
	}
}
