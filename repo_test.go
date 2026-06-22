package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRepoName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		repoPath   string
		repoRemote string
		want       string
	}{
		{name: "path", repoPath: "/tmp/checkout-api", want: "checkout-api"},
		{name: "https remote", repoRemote: "https://github.com/prilog-ai/platform.git", want: "platform"},
		{name: "ssh remote", repoRemote: "git@github.com:prilog-ai/platform.git", want: "platform"},
		{name: "fallback", want: defaultProjectName},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := repoName(tc.repoPath, tc.repoRemote); got != tc.want {
				t.Fatalf("repoName(%q, %q) = %q, want %q", tc.repoPath, tc.repoRemote, got, tc.want)
			}
		})
	}
}

func TestPromptProjectNameNonTerminalUsesDefault(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	c := newCLI(defaultAPIURL, ".", strings.NewReader("custom\n"), &out, nil)
	got, err := c.promptProjectName("repo-name")
	if err != nil {
		t.Fatalf("promptProjectName returned error: %v", err)
	}
	if got != "repo-name" {
		t.Fatalf("promptProjectName = %q, want default", got)
	}
	if out.Len() != 0 {
		t.Fatalf("non-terminal prompt wrote output: %q", out.String())
	}
}
