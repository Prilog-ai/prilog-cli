package main

import "time"

type globalConfig struct {
	APIURL       string    `json:"api_url"`
	DashboardURL string    `json:"dashboard_url,omitempty"`
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	UserID       string    `json:"user_id,omitempty"`
	CLISessionID string    `json:"cli_session_id,omitempty"`
}

type repoConfig struct {
	APIURL               string `json:"api_url"`
	DashboardURL         string `json:"dashboard_url,omitempty"`
	ProjectOnboardingURL string `json:"project_onboarding_url,omitempty"`
	OrganizationID       string `json:"organization_id,omitempty"`
	ProjectID            string `json:"project_id,omitempty"`
	RepoPath             string `json:"repo_path,omitempty"`
	RepoRemote           string `json:"repo_remote,omitempty"`
	IngestURL            string `json:"ingest_url,omitempty"`
	OTELLogsURL          string `json:"otel_logs_url,omitempty"`
	OTELTracesURL        string `json:"otel_traces_url,omitempty"`
}

type authStartResponse struct {
	DeviceCode string `json:"device_code"`
	AuthURL    string `json:"auth_url"`
	ExpiresIn  int    `json:"expires_in"`
	Interval   int    `json:"interval"`
}

type authPollResponse struct {
	Status           string `json:"status"`
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	OrganizationID   string `json:"organization_id"`
	UserID           string `json:"user_id"`
	CLISessionID     string `json:"cli_session_id"`
	DashboardURL     string `json:"dashboard_url"`
}

type authRefreshResponse struct {
	AccessToken    string `json:"access_token"`
	TokenType      string `json:"token_type"`
	ExpiresIn      int    `json:"expires_in"`
	OrganizationID string `json:"organization_id"`
	UserID         string `json:"user_id"`
	ProjectID      string `json:"project_id"`
}

type projectInitResponse struct {
	Status                   string `json:"status"`
	ProjectID                string `json:"project_id"`
	OrganizationID           string `json:"organization_id"`
	RepoPath                 string `json:"repo_path"`
	RepoRemote               string `json:"repo_remote"`
	DashboardURL             string `json:"dashboard_url"`
	ProjectOnboardingURL     string `json:"project_onboarding_url"`
	ProjectOnboardingAuthURL string `json:"project_onboarding_auth_url"`
	IngestURL                string `json:"ingest_url"`
	OTELLogsURL              string `json:"otel_logs_url"`
	OTELTracesURL            string `json:"otel_traces_url"`
	Created                  bool   `json:"created"`
}

type ingestResponse struct {
	Accepted int    `json:"accepted"`
	Skipped  int    `json:"skipped"`
	Provider string `json:"provider"`
	Signal   string `json:"signal"`
}

type errorLogsResponse struct {
	Errors     []errorLog `json:"errors"`
	TotalCount int        `json:"total_count"`
}

type errorLog struct {
	ID                 string         `json:"id"`
	Message            string         `json:"message"`
	Severity           string         `json:"severity"`
	Status             string         `json:"status"`
	Timestamp          time.Time      `json:"timestamp"`
	RepositoryPath     string         `json:"repository_path"`
	ResolutionURL      *string        `json:"resolution_url"`
	ResolutionMetadata map[string]any `json:"resolution_metadata"`
}

type fixResponse struct {
	Status    string `json:"status"`
	ID        string `json:"id"`
	ReviewURL string `json:"review_url"`
}

type diffResponse struct {
	ID        string `json:"id"`
	Diff      string `json:"diff"`
	ReviewURL string `json:"review_url"`
}

type statusResponse struct {
	User         statusUser         `json:"user"`
	Organization statusOrganization `json:"organization"`
	Project      statusProject      `json:"project"`
	Counts       statusCounts       `json:"counts"`
	DashboardURL string             `json:"dashboard_url"`
}

type statusUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

type statusOrganization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type statusProject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type statusCounts struct {
	Logs     int            `json:"logs"`
	Fixes    int            `json:"fixes"`
	ByStatus map[string]int `json:"by_status"`
}
