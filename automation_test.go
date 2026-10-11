package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestJSONAutomationPinsProjectAndAPI(t *testing.T) {
	t.Setenv("PRILOG_CONFIG_DIR", t.TempDir())
	projectID := "11111111-1111-1111-1111-111111111111"
	var path, project string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		project = r.Header.Get("X-Prilog-Project-ID")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":[],"total_count":0}`))
	}))
	defer server.Close()
	if err := saveGlobalConfig(globalConfig{APIURL: "https://wrong-endpoint.example/api", AccessToken: "test-token", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := runWithIO([]string{"--api-url", server.URL, "--json", "--project-id", projectID, "list", "pending"}, nil, &output, nil); err != nil {
		t.Fatal(err)
	}
	var response errorLogsResponse
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatalf("invalid JSON: %s %v", output.String(), err)
	}
	if path != "/errors" || project != projectID {
		t.Fatalf("request = %s, project %s", path, project)
	}
}

func TestJSONMissingAuthenticationNeverOpensBrowser(t *testing.T) {
	t.Setenv("PRILOG_CONFIG_DIR", t.TempDir())
	var output bytes.Buffer
	opened := false
	err := runWithIO([]string{"--json", "status"}, nil, &output, func(string) error { opened = true; return nil })
	if err == nil || !strings.Contains(err.Error(), "prilog login") || opened || output.Len() != 0 {
		t.Fatalf("interactive automation: %v, opened=%v, output=%s", err, opened, output.String())
	}
}

func TestAutomationRejectsInvalidProject(t *testing.T) {
	var output bytes.Buffer
	err := runWithIO([]string{"--project-id", "not-a-project", "status"}, nil, &output, nil)
	if err == nil || !strings.Contains(err.Error(), "UUID") {
		t.Fatalf("invalid project accepted: %v", err)
	}
}

func TestSubcommandHelpNeverAuthenticatesOrMutates(t *testing.T) {
	t.Setenv("PRILOG_CONFIG_DIR", t.TempDir())
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(http.StatusInternalServerError) }))
	defer server.Close()
	if err := saveGlobalConfig(globalConfig{APIURL: server.URL, AccessToken: "test-token", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"create-project", "use-project", "fix", "diff", "pr", "login", "logout", "add-repo", "status", "projects", "ingest", "issue"} {
		for _, help := range []string{"--help", "-h"} {
			var output bytes.Buffer
			opened := false
			if err := runWithIO([]string{"--json", command, help}, nil, &output, func(string) error { opened = true; return nil }); err != nil {
				t.Fatalf("%s %s: %v", command, help, err)
			}
			if !strings.Contains(output.String(), "Usage:") || opened {
				t.Fatalf("%s %s did not return safe help", command, help)
			}
		}
	}
	if calls != 0 {
		t.Fatalf("help made %d network requests", calls)
	}
}
