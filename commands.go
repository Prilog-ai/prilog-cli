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

func (c *cli) init() error {
	repoPath, repoRemote := c.repoMetadata()
	projectName, err := c.promptProjectName(repoName(repoPath, repoRemote))
	if err != nil {
		return err
	}
	if err := c.ensureAuth(); err != nil {
		return err
	}

	payload := map[string]string{
		"repo_path":   repoPath,
		"repo_remote": repoRemote,
		"repo_name":   projectName,
	}
	var response projectInitResponse
	if err := c.doJSON(context.Background(), http.MethodPost, "/cli/projects/init", payload, true, &response); err != nil {
		return err
	}

	repoCfg := repoConfig{
		APIURL:               c.apiURL,
		DashboardURL:         response.DashboardURL,
		ProjectOnboardingURL: response.ProjectOnboardingURL,
		OrganizationID:       response.OrganizationID,
		ProjectID:            response.ProjectID,
		RepoPath:             response.RepoPath,
		RepoRemote:           response.RepoRemote,
		IngestURL:            response.IngestURL,
		OTELLogsURL:          response.OTELLogsURL,
		OTELTracesURL:        response.OTELTracesURL,
	}
	if err := saveRepoConfig(c.root, repoCfg); err != nil {
		return err
	}

	action := "linked"
	if response.Created {
		action = "created"
	}
	c.printf("Prilog project %s: %s\n", action, response.ProjectID)

	openURL := firstNonEmpty(response.ProjectOnboardingAuthURL, response.ProjectOnboardingURL, response.DashboardURL)
	if openURL != "" {
		c.println("Opening project onboarding:")
		c.println(firstNonEmpty(response.ProjectOnboardingURL, openURL))
		_ = c.openURL(openURL)
	}
	return nil
}

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
	fmt.Fprintf(w, "api_url\t%s\n", firstNonEmpty(repoCfg.APIURL, global.APIURL, c.apiURL))
	fmt.Fprintf(w, "authenticated\t%t\n", global.AccessToken != "")
	fmt.Fprintf(w, "token_expires_at\t%s\n", formatTime(global.ExpiresAt))
	fmt.Fprintf(w, "organization_id\t%s\n", repoCfg.OrganizationID)
	fmt.Fprintf(w, "project_id\t%s\n", repoCfg.ProjectID)
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
	repoCfg, err := loadRepoConfig(c.root)
	if err != nil || repoCfg.ProjectID == "" {
		return errors.New("repository is not initialized; run `prilog init` first")
	}

	body, filename, err := readIngestPayload(args, c.stdin)
	if err != nil {
		return err
	}

	values := url.Values{}
	values.Set("filename", filepath.Base(filename))
	if repoCfg.RepoPath != "" {
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
	c.printf("Accepted %d %s record(s), skipped %d.\n", response.Accepted, response.Signal, response.Skipped)
	return nil
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
	case "pending", "processing", "completed", "failed":
		return status, nil
	default:
		return "", fmt.Errorf("unsupported list filter %q. Supported filters: all, pending, processing, completed, failed", status)
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
	c.println(prURL)
	return nil
}
