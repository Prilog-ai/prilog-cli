package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (c *cli) config(args []string) error {
	if len(args) == 3 && args[0] == "set" && args[1] == "api-url" {
		return c.setAPIURL(args[2])
	}
	if len(args) == 1 && args[0] == "path" {
		return c.printConfigPaths()
	}
	if len(args) != 0 {
		return errors.New("config does not accept public arguments; run `prilog config`")
	}
	return c.printConfig()
}

func (c *cli) setAPIURL(rawURL string) error {
	global, _ := loadGlobalConfig()
	global.APIURL = trimConfiguredURL(rawURL)
	if err := saveGlobalConfig(global); err != nil {
		return err
	}

	if repoCfg, err := loadRepoConfig(c.root); err == nil {
		repoCfg.APIURL = global.APIURL
		if err := saveRepoConfig(c.root, repoCfg); err != nil {
			return err
		}
	}

	c.println("API URL updated:", global.APIURL)
	return nil
}

func (c *cli) printConfigPaths() error {
	globalPath, _ := globalConfigPath()
	c.println("Global:", globalPath)
	c.println("Repo:", filepath.Join(c.root, configDir, configFile))
	return nil
}

func (c *cli) printConfig() error {
	global, _ := loadGlobalConfig()
	repoCfg, _ := loadRepoConfig(c.root)

	w := newTabWriter(c.stdout)
	fmt.Fprintln(w, "KEY\tVALUE")
	fmt.Fprintf(w, "api_url\t%s\n", c.apiURL)
	fmt.Fprintf(w, "authenticated\t%t\n", global.AccessToken != "")
	fmt.Fprintf(w, "token_expires_at\t%s\n", formatTime(global.ExpiresAt))
	fmt.Fprintf(w, "organization_id\t%s\n", firstNonEmpty(global.OrganizationID, repoCfg.OrganizationID))
	fmt.Fprintf(w, "project_id\t%s\n", c.requestProjectID())
	if c.requestProjectID() == global.ProjectID {
		fmt.Fprintf(w, "project_name\t%s\n", global.ProjectName)
	}
	fmt.Fprintf(w, "repo_path\t%s\n", repoCfg.RepoPath)
	fmt.Fprintf(w, "dashboard_url\t%s\n", firstNonEmpty(repoCfg.DashboardURL, global.DashboardURL))
	fmt.Fprintf(w, "project_onboarding_url\t%s\n", repoCfg.ProjectOnboardingURL)
	return w.Flush()
}

func (c *cli) status() error {
	if err := c.ensureAuth(); err != nil {
		return err
	}

	var response statusResponse
	if err := c.doJSON(context.Background(), http.MethodGet, "/cli/status", nil, true, &response); err != nil {
		return err
	}

	if c.jsonOutput {
		return c.writeJSON(response)
	}
	w := newTabWriter(c.stdout)
	fmt.Fprintln(w, "KEY\tVALUE")
	fmt.Fprintf(w, "user\t%s\n", statusUserLabel(response.User))
	fmt.Fprintf(w, "organization\t%s\n", statusEntityLabel(response.Organization.Name, response.Organization.ID))
	fmt.Fprintf(w, "project\t%s\n", statusEntityLabel(response.Project.Name, response.Project.ID))
	fmt.Fprintf(w, "logs\t%d\n", response.Counts.Logs)
	fmt.Fprintf(w, "fixes\t%d\n", response.Counts.Fixes)
	fmt.Fprintf(w, "log_statuses\t%s\n", statusCountsLabel(response.Counts.ByStatus))
	if response.DashboardURL != "" {
		fmt.Fprintf(w, "dashboard_url\t%s\n", response.DashboardURL)
	}
	return w.Flush()
}

func (c *cli) ingest(args []string) error {
	if err := c.ensureAuth(); err != nil {
		return err
	}
	repoCfg, _ := loadRepoConfig(c.root)
	if c.requestProjectID() == "" {
		return errors.New("no project selected; run `prilog use-project <project-name>` or `prilog add-repo` first")
	}

	body, filename, err := readIngestPayload(args, c.stdin)
	if err != nil {
		return err
	}

	values := url.Values{}
	values.Set("filename", filepath.Base(filename))
	if repoCfg.RepoPath != "" && repoCfg.ProjectID == c.requestProjectID() {
		values.Set("repo_path", repoCfg.RepoPath)
	}

	path := "/cli/ingest"
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var response ingestResponse
	if err := c.doRaw(context.Background(), http.MethodPost, path, body, true, &response); err != nil {
		return err
	}
	if c.jsonOutput {
		return c.writeJSON(response)
	}
	c.printf("Accepted %d %s record(s), skipped %d.\n", response.Accepted, response.Signal, response.Skipped)
	return nil
}

func (c *cli) issue(args []string) error {
	payload, err := githubIssueImportPayload(args)
	if err != nil {
		return err
	}
	if err := c.ensureAuth(); err != nil {
		return err
	}
	if c.requestProjectID() == "" {
		return errors.New("no project selected; run `prilog use-project <project-name>` or `prilog add-repo` first")
	}

	var response githubIssueImportResponse
	if err := c.doJSON(context.Background(), http.MethodPost, "/cli/github/issues/import", payload, true, &response); err != nil {
		return err
	}
	if c.jsonOutput {
		return c.writeJSON(response)
	}
	c.printf("Imported %s#%d: %s\n", response.Repository, response.Number, response.Title)
	c.println("Error ID:", response.ID)
	c.println("Next: prilog fix", response.ID)
	if response.ReviewURL != "" {
		c.println("Review:", response.ReviewURL)
	}
	return nil
}

func githubIssueImportPayload(args []string) (map[string]string, error) {
	if len(args) != 2 || args[0] != "import" {
		return nil, errors.New("usage: prilog issue import <github-issue-url|owner/repo#number>")
	}
	issueRef := strings.TrimSpace(args[1])
	if issueRef == "" {
		return nil, errors.New("GitHub issue URL or owner/repo#number is required")
	}
	return map[string]string{"issue": issueRef}, nil
}

func readIngestPayload(args []string, stdin io.Reader) ([]byte, string, error) {
	if len(args) == 0 {
		body, err := io.ReadAll(stdin)
		return body, "stdin", err
	}

	filename := args[0]
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, "", err
	}
	return body, filename, nil
}

func (c *cli) list(status string) error {
	status, err := normalizeListStatus(status)
	if err != nil {
		return err
	}
	if err := c.ensureAuth(); err != nil {
		return err
	}

	values := url.Values{}
	values.Set("page_size", "100")
	if status != "" {
		values.Set("status", status)
	}

	var response errorLogsResponse
	if err := c.doJSON(context.Background(), http.MethodGet, "/errors?"+values.Encode(), nil, true, &response); err != nil {
		return err
	}
	if c.jsonOutput {
		return c.writeJSON(response)
	}
	if len(response.Errors) == 0 {
		c.println("No logs found.")
		return nil
	}

	w := newTabWriter(c.stdout)
	fmt.Fprintln(w, "ID\tSTATUS\tSEVERITY\tWHEN\tMESSAGE")
	for _, item := range response.Errors {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", item.ID, item.Status, item.Severity, item.Timestamp.Format(time.RFC3339), truncate(item.Message, 100))
	}
	return w.Flush()
}

func normalizeListStatus(status string) (string, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "", "all":
		return "", nil
	case "pending", "processing", "completed", "failed", "skipped":
		return status, nil
	default:
		return "", fmt.Errorf("unsupported list filter %q. Supported filters: all, pending, processing, completed, failed, skipped", status)
	}
}

func (c *cli) fix(id string) error {
	if err := c.ensureAuth(); err != nil {
		return err
	}

	var response fixResponse
	if err := c.doJSON(context.Background(), http.MethodPost, "/cli/errors/"+id+"/fix", map[string]string{}, true, &response); err != nil {
		return err
	}

	if c.jsonOutput {
		return c.writeJSON(response)
	}
	c.println("Analysis queued:", response.ID)
	if response.ReviewURL != "" {
		c.println(response.ReviewURL)
	}
	return nil
}

func (c *cli) diff(id string) error {
	if err := c.ensureAuth(); err != nil {
		return err
	}

	var response diffResponse
	if err := c.doJSON(context.Background(), http.MethodGet, "/cli/errors/"+id+"/diff", nil, true, &response); err != nil {
		return err
	}
	if c.jsonOutput {
		return c.writeJSON(response)
	}
	if strings.TrimSpace(response.Diff) == "" {
		c.println("No diff available.")
		return nil
	}

	c.println(response.Diff)
	return nil
}

func (c *cli) pr(id string) error {
	if err := c.ensureAuth(); err != nil {
		return err
	}

	var response errorLog
	if err := c.doJSON(context.Background(), http.MethodPost, "/fixes/"+id+"/decision", map[string]string{"action": "pr"}, true, &response); err != nil {
		return err
	}

	prURL := extractPRURL(response)
	if prURL == "" {
		return errors.New("pull request was created but no PR URL was returned")
	}
	if c.jsonOutput {
		return c.writeJSON(response)
	}
	c.println(prURL)
	return nil
}
