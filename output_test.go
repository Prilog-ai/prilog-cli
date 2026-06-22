package main

import "testing"

func TestOutputHelpers(t *testing.T) {
	t.Parallel()

	if got := truncate(" one   two three ", 7); got != "one ..." {
		t.Fatalf("truncate = %q", got)
	}
	if got := firstNonEmpty("", "  ", "api"); got != "api" {
		t.Fatalf("firstNonEmpty = %q", got)
	}
	if got := statusCountsLabel(map[string]int{"pending": 2, "completed": 1}); got != "pending=2 processing=0 completed=1 failed=0" {
		t.Fatalf("statusCountsLabel = %q", got)
	}
}

func TestExtractPRURL(t *testing.T) {
	t.Parallel()

	direct := "https://github.com/prilog-ai/platform/pull/1"
	if got := extractPRURL(errorLog{ResolutionURL: &direct}); got != direct {
		t.Fatalf("direct PR URL = %q", got)
	}

	nested := errorLog{ResolutionMetadata: map[string]any{
		"actions": map[string]any{
			"pr": map[string]any{"url": "https://github.com/prilog-ai/platform/pull/2"},
		},
	}}
	if got := extractPRURL(nested); got != "https://github.com/prilog-ai/platform/pull/2" {
		t.Fatalf("nested PR URL = %q", got)
	}
}
