package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"
)

func (c *cli) loadProjects() (projectsResponse, error) {
	var response projectsResponse
	if err := c.ensureAuth(); err != nil {
		return response, err
	}
	err := c.doJSON(context.Background(), http.MethodGet, "/projects", nil, true, &response)
	if response.Projects == nil {
		response.Projects = []projectResponse{}
	}
	return response, err
}

func (c *cli) projects() error {
	response, err := c.loadProjects()
	if err != nil {
		return err
	}
	selected := c.requestProjectID()
	if selected == "" {
		selected = response.ActiveProjectID
	}
	response.ActiveProjectID = ""
	for _, project := range response.Projects {
		if project.ID == selected {
			response.ActiveProjectID = project.ID
		}
	}
	if c.jsonOutput {
		return c.writeJSON(response)
	}
	if len(response.Projects) == 0 {
		c.println("No projects. Run `prilog create-project <project-name>` or `prilog add-repo`.")
		return nil
	}
	w := newTabWriter(c.stdout)
	fmt.Fprintln(w, "CURRENT\tNAME\tID")
	for _, project := range response.Projects {
		marker := ""
		if project.ID == response.ActiveProjectID {
			marker = "*"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", marker, project.Name, project.ID)
	}
	return w.Flush()
}

func projectArgument(command string, args []string) (string, error) {
	name := strings.TrimSpace(strings.Join(args, " "))
	if name == "" {
		return "", fmt.Errorf("%s requires a project name", command)
	}
	return name, nil
}

func (c *cli) createProject(args []string) error {
	if len(args) == 0 && isTerminal(c.stdin) && !c.jsonOutput {
		c.printf("Project name: ")
		value, err := bufio.NewReader(c.stdin).ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}
		args = []string{value}
	}
	name, err := projectArgument("create-project", args)
	if err != nil {
		return err
	}
	if utf8.RuneCountInString(name) > 50 {
		return errors.New("project name must be 50 characters or fewer")
	}
	if err := c.ensureAuth(); err != nil {
		return err
	}
	var project projectResponse
	if err := c.doJSON(context.Background(), http.MethodPost, "/projects", map[string]string{"name": name}, true, &project); err != nil {
		return err
	}
	if err := c.saveProjectSelection(project); err != nil {
		return fmt.Errorf("project %s exists but its default selection could not be saved; retry `prilog use-project %s`: %w", project.ID, project.ID, err)
	}
	if c.jsonOutput {
		return c.writeJSON(project)
	}
	c.printf("Created and selected project: %s (%s)\n", project.Name, project.ID)
	return nil
}

func (c *cli) useProject(args []string) error {
	name, err := projectArgument("use-project", args)
	if err != nil {
		return err
	}
	response, err := c.loadProjects()
	if err != nil {
		return err
	}
	matches := []projectResponse{}
	for _, project := range response.Projects {
		if project.ID == name {
			matches = []projectResponse{project}
			break
		}
		if strings.EqualFold(project.Name, name) {
			matches = append(matches, project)
		}
	}
	if len(matches) == 0 {
		return fmt.Errorf("project %q was not found; run `prilog projects`", name)
	}
	if len(matches) > 1 {
		return fmt.Errorf("project name %q is ambiguous; run `prilog projects` and use the project id", name)
	}
	var project projectResponse
	if err := c.doJSON(context.Background(), http.MethodPost, "/projects/active", map[string]string{"project_id": matches[0].ID}, true, &project); err != nil {
		return err
	}
	if err := c.saveProjectSelection(project); err != nil {
		return fmt.Errorf("project %s exists but its default selection could not be saved; retry `prilog use-project %s`: %w", project.ID, project.ID, err)
	}
	if c.jsonOutput {
		return c.writeJSON(project)
	}
	c.printf("Selected project: %s (%s)\n", project.Name, project.ID)
	return nil
}

func (c *cli) saveProjectSelection(project projectResponse) error {
	if !validProjectID(project.ID) {
		return errors.New("Prilog returned an invalid project id")
	}
	global, err := loadGlobalConfig()
	if err != nil {
		return err
	}
	global.APIURL = c.apiURL
	global.ProjectID = project.ID
	global.ProjectName = project.Name
	if project.OrganizationID != "" {
		global.OrganizationID = project.OrganizationID
	}
	if err := saveGlobalConfig(global); err != nil {
		return err
	}
	c.projectID = project.ID
	return nil
}

func (c *cli) addRepo() error {
	if err := c.ensureAuth(); err != nil {
		return err
	}
	var response addRepoResponse
	if err := c.doJSON(context.Background(), http.MethodPost, "/cli/repos/start", map[string]string{}, true, &response); err != nil {
		return err
	}
	if response.BrowserURL == "" {
		return errors.New("Prilog did not return a repository setup URL")
	}
	if err := c.saveProjectSelection(response.Project); err != nil {
		return err
	}
	if c.jsonOutput {
		return c.writeJSON(response)
	}
	c.printf("Adding a GitHub repository to %s.\n", response.Project.Name)
	c.println("Opening Prilog repository setup:", response.BrowserURL)
	if err := c.openURL(response.BrowserURL); err != nil {
		return fmt.Errorf("could not open the browser; open the setup URL above: %w", err)
	}
	return nil
}
