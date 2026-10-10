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

const testProjectA = "11111111-1111-1111-1111-111111111111"
const testProjectB = "22222222-2222-2222-2222-222222222222"

func projectCLI(t *testing.T, handler http.HandlerFunc) *cli {
	t.Helper()
	t.Setenv("PRILOG_CONFIG_DIR", t.TempDir())
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	if err := saveGlobalConfig(globalConfig{APIURL: server.URL, AccessToken: "test-token", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	return newCLI(server.URL, t.TempDir(), nil, &bytes.Buffer{}, nil)
}

func TestCreateAndUseProjectPersistAcrossDirectories(t *testing.T) {
	var calls []string
	c := projectCLI(t, func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/projects":
			if r.Method == "POST" {
				var body map[string]string
				_ = json.NewDecoder(r.Body).Decode(&body)
				if body["name"] != "Payments" {
					t.Error("wrong project name")
				}
				_, _ = w.Write([]byte(`{"id":"` + testProjectA + `","name":"Payments","organization_id":"org"}`))
			} else {
				_, _ = w.Write([]byte(`{"projects":[{"id":"` + testProjectA + `","name":"Payments"},{"id":"` + testProjectB + `","name":"Worker"}],"total_count":2}`))
			}
		case "/projects/active":
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["project_id"] != testProjectB {
				t.Error("wrong selected project")
			}
			_, _ = w.Write([]byte(`{"id":"` + testProjectB + `","name":"Worker","organization_id":"org"}`))
		case "/cli/status":
			if r.Header.Get("X-Prilog-Project-ID") != testProjectB {
				t.Error("selection lost across directories")
			}
			_, _ = w.Write([]byte(`{"project":{"id":"` + testProjectB + `","name":"Worker"}}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})
	c.jsonOutput = true
	if err := c.createProject([]string{"Payments"}); err != nil {
		t.Fatal(err)
	}
	global, _ := loadGlobalConfig()
	if global.ProjectID != testProjectA || global.OrganizationID != "org" {
		t.Fatal("creation did not select project")
	}
	if err := c.useProject([]string{"worker"}); err != nil {
		t.Fatal(err)
	}
	next := newCLI(c.apiURL, t.TempDir(), nil, &bytes.Buffer{}, nil)
	next.jsonOutput = true
	if err := saveRepoConfig(next.root, repoConfig{ProjectID: testProjectA}); err != nil {
		t.Fatal(err)
	}
	if err := next.status(); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 4 {
		t.Fatalf("unexpected command calls: %v", calls)
	}
}

func TestProjectsIncludesAllAndMarksCurrent(t *testing.T) {
	c := projectCLI(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"projects":[{"id":"` + testProjectA + `","name":"Payments"},{"id":"` + testProjectB + `","name":"Worker"}],"total_count":2,"active_project_id":"` + testProjectB + `"}`))
	})
	var output bytes.Buffer
	c.stdout = &output
	c.jsonOutput = true
	if err := c.projects(); err != nil {
		t.Fatal(err)
	}
	var result projectsResponse
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Projects) != 2 || result.ActiveProjectID != testProjectB {
		t.Fatalf("project listing: %+v", result)
	}
}

func TestUseProjectRejectsAmbiguousOrMissingNames(t *testing.T) {
	var selections int
	c := projectCLI(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			selections++
			return
		}
		_, _ = w.Write([]byte(`{"projects":[{"id":"` + testProjectA + `","name":"Payments"},{"id":"` + testProjectB + `","name":"payments"}]}`))
	})
	for _, name := range []string{"Payments", "missing"} {
		if err := c.useProject([]string{name}); err == nil {
			t.Fatalf("invalid selection %q accepted", name)
		}
	}
	if selections != 0 {
		t.Fatal("invalid selection sent a write request")
	}
}

func TestAddRepoStartsSelectedProjectAndOpensPrilog(t *testing.T) {
	c := projectCLI(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cli/repos/start" || r.Method != "POST" || r.Header.Get("X-Prilog-Project-ID") != testProjectB {
			t.Error("incorrect repository handoff request")
		}
		_, _ = w.Write([]byte(`{"project":{"id":"` + testProjectB + `","name":"Worker"},"browser_url":"https://app.prilog.ai/cli/add-repo?state=one-time"}`))
	})
	c.projectID = testProjectB
	opened := ""
	c.openURL = func(url string) error { opened = url; return nil }
	if err := c.addRepo(); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(opened, "https://app.prilog.ai/cli/add-repo?") {
		t.Fatal("Prilog browser handoff was not opened")
	}
	c.jsonOutput = true
	opened = ""
	c.stdout = &bytes.Buffer{}
	if err := c.addRepo(); err != nil {
		t.Fatal(err)
	}
	if opened != "" {
		t.Fatal("JSON mode launched a browser")
	}
}

func TestProjectCommandValidationAndRemovedInit(t *testing.T) {
	c := newCLI(defaultAPIURL, t.TempDir(), strings.NewReader(""), &bytes.Buffer{}, nil)
	for _, args := range [][]string{{"init"}, {"create-project"}, {"create-project", strings.Repeat("a", 51)}, {"use-project"}, {"projects", "extra"}, {"add-repo", "extra"}} {
		if err := c.dispatch(args); err == nil {
			t.Fatalf("invalid command accepted: %v", args)
		}
	}
}
