package main

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestSingleIDArg(t *testing.T) {
	t.Parallel()

	id, err := singleIDArg("fix", []string{":abc-123"})
	if err != nil {
		t.Fatalf("singleIDArg returned error: %v", err)
	}
	if id != "abc-123" {
		t.Fatalf("singleIDArg id = %q, want %q", id, "abc-123")
	}

	if _, err := singleIDArg("fix", nil); err == nil || !strings.Contains(err.Error(), "requires an error id") {
		t.Fatalf("missing id error = %v", err)
	}
	if _, err := singleIDArg("fix", []string{"one", "two"}); err == nil || !strings.Contains(err.Error(), "exactly one") {
		t.Fatalf("extra id error = %v", err)
	}
}

func TestPrintUsage(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	printUsage(&out)
	for _, want := range []string{"prilog init", "prilog status", "prilog pr <id>"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("usage missing %q:\n%s", want, out.String())
		}
	}
	if strings.Contains(out.String(), "api-url") {
		t.Fatalf("usage exposed internal API override:\n%s", out.String())
	}
}

func TestPrintVersion(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	printVersion(&out)
	if got := strings.TrimSpace(out.String()); got != "prilog dev" {
		t.Fatalf("version = %q, want %q", got, "prilog dev")
	}
}

func TestRunWithIOHelp(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	if err := runWithIO([]string{"help"}, nil, &out, nil); err != nil {
		t.Fatalf("runWithIO help returned error: %v", err)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Fatalf("help output missing usage: %q", out.String())
	}
}

func TestRunWithIORejectsUnexpectedArgs(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	err := runWithIO([]string{"status", "extra"}, nil, &out, nil)
	if err == nil || !strings.Contains(err.Error(), "status does not accept arguments") {
		t.Fatalf("unexpected status arg error = %v", err)
	}
}

func TestDecodeAPIError(t *testing.T) {
	t.Parallel()

	err := decodeAPIError(http.StatusForbidden, []byte(`{"error":"write_required"}`))
	var apiErr apiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("decodeAPIError type = %T, want apiError", err)
	}
	if apiErr.StatusCode != http.StatusForbidden || apiErr.Message != "write_required" {
		t.Fatalf("api error = %+v", apiErr)
	}

	err = decodeAPIError(http.StatusBadGateway, []byte("upstream failed"))
	if err.Error() != "api returned HTTP 502: upstream failed" {
		t.Fatalf("plain api error = %q", err.Error())
	}
}
