package main

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestLogoutClearsLoginAndPreservesRepositorySettings(t *testing.T) {
	t.Setenv("PRILOG_CONFIG_DIR", t.TempDir())
	root := t.TempDir()
	if err := saveGlobalConfig(globalConfig{APIURL: "http://localhost:3333/api", DashboardURL: "http://localhost:5173", AccessToken: "test-access", RefreshToken: "test-refresh", TokenType: "Bearer", UserID: "user", CLISessionID: "session", OrganizationID: "organization", ProjectID: testProjectA, ProjectName: "Payments", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	if err := saveRepoConfig(root, repoConfig{ProjectID: "project", OrganizationID: "organization"}); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	c := newCLI("http://localhost:3333/api", root, nil, &output, nil)
	if err := c.dispatch([]string{"logout"}); err != nil {
		t.Fatal(err)
	}
	config, err := loadGlobalConfig()
	if err != nil {
		t.Fatal(err)
	}
	if config.AccessToken != "" || config.RefreshToken != "" || config.TokenType != "" || config.UserID != "" || config.CLISessionID != "" || config.ProjectID != "" || config.ProjectName != "" || config.OrganizationID != "" || !config.ExpiresAt.IsZero() {
		t.Fatal("saved login was not cleared")
	}
	if config.APIURL != "http://localhost:3333/api" || config.DashboardURL != "http://localhost:5173" {
		t.Fatal("endpoint preferences changed")
	}
	repository, err := loadRepoConfig(root)
	if err != nil || repository.ProjectID != "project" || repository.OrganizationID != "organization" {
		t.Fatal("repository selection changed")
	}
	c.jsonOutput = true
	if err := c.status(); err == nil || !strings.Contains(err.Error(), "prilog login") {
		t.Fatalf("signed-out command did not require login: %v", err)
	}
}

func TestLogoutWithoutLoginIsIdempotent(t *testing.T) {
	t.Setenv("PRILOG_CONFIG_DIR", t.TempDir())
	var output bytes.Buffer
	c := newCLI("http://localhost:3333/api", t.TempDir(), nil, &output, nil)
	for range 2 {
		if err := c.dispatch([]string{"logout"}); err != nil {
			t.Fatal(err)
		}
	}
	if err := c.dispatch([]string{"logout", "extra"}); err == nil {
		t.Fatal("extra logout arguments accepted")
	}
}
