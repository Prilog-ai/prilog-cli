package main

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"
)

func newTabWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
}

func cleanID(id string) string {
	return strings.TrimPrefix(strings.TrimSpace(id), ":")
}

func truncate(value string, limit int) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return value[:limit]
	}
	return value[:limit-3] + "..."
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func extractPRURL(response errorLog) string {
	if response.ResolutionURL != nil && strings.TrimSpace(*response.ResolutionURL) != "" {
		return strings.TrimSpace(*response.ResolutionURL)
	}
	if response.ResolutionMetadata == nil {
		return ""
	}
	if value, ok := response.ResolutionMetadata["pr_url"].(string); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}

	actions, ok := response.ResolutionMetadata["actions"].(map[string]any)
	if !ok {
		return ""
	}
	pr, ok := actions["pr"].(map[string]any)
	if !ok {
		return ""
	}
	if value, ok := pr["url"].(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func statusUserLabel(user statusUser) string {
	name := strings.TrimSpace(user.Name)
	email := strings.TrimSpace(user.Email)
	if name != "" && email != "" {
		if user.Role != "" {
			return fmt.Sprintf("%s <%s> (%s)", name, email, user.Role)
		}
		return fmt.Sprintf("%s <%s>", name, email)
	}
	if email != "" {
		if user.Role != "" {
			return fmt.Sprintf("%s (%s)", email, user.Role)
		}
		return email
	}
	return user.ID
}

func statusEntityLabel(name, id string) string {
	name = strings.TrimSpace(name)
	id = strings.TrimSpace(id)
	if name != "" && id != "" {
		return fmt.Sprintf("%s (%s)", name, id)
	}
	if name != "" {
		return name
	}
	return id
}

func statusCountsLabel(counts map[string]int) string {
	if counts == nil {
		counts = map[string]int{}
	}

	statuses := []string{"pending", "processing", "completed", "failed"}
	parts := make([]string, 0, len(statuses))
	for _, status := range statuses {
		parts = append(parts, fmt.Sprintf("%s=%d", status, counts[status]))
	}
	return strings.Join(parts, " ")
}
