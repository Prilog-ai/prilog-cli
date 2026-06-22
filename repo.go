package main

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func repoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	out, err := exec.Command("git", "-C", cwd, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return cwd, nil
	}
	root := strings.TrimSpace(string(out))
	if root == "" {
		return cwd, nil
	}
	return root, nil
}

func (c *cli) repoMetadata() (string, string) {
	out, err := exec.Command("git", "-C", c.root, "config", "--get", "remote.origin.url").Output()
	if err != nil {
		return c.root, ""
	}
	return c.root, strings.TrimSpace(string(out))
}

func repoName(repoPath, repoRemote string) string {
	if repoPath != "" {
		if base := filepath.Base(repoPath); base != "." && base != string(filepath.Separator) {
			return strings.TrimSuffix(base, ".git")
		}
	}
	if repoRemote != "" {
		if parsed, err := url.Parse(repoRemote); err == nil && parsed.Path != "" {
			return strings.TrimSuffix(filepath.Base(parsed.Path), ".git")
		}
		if idx := strings.LastIndex(repoRemote, "/"); idx >= 0 && idx < len(repoRemote)-1 {
			return strings.TrimSuffix(repoRemote[idx+1:], ".git")
		}
		if idx := strings.LastIndex(repoRemote, ":"); idx >= 0 && idx < len(repoRemote)-1 {
			return strings.TrimSuffix(repoRemote[idx+1:], ".git")
		}
	}
	return defaultProjectName
}

func (c *cli) promptProjectName(defaultName string) (string, error) {
	defaultName = strings.TrimSpace(defaultName)
	if defaultName == "" {
		defaultName = defaultProjectName
	}
	if !isTerminal(c.stdin) {
		return defaultName, nil
	}

	fmt.Fprintf(c.stdout, "Project name [%s]: ", defaultName)
	reader := bufio.NewReader(c.stdin)
	value, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultName, nil
	}
	return value, nil
}

func isTerminal(reader io.Reader) bool {
	file, ok := reader.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

const defaultProjectName = "CLI Project"
