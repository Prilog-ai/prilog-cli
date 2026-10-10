package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type browserOpener func(string) error

type cli struct {
	apiURL     string
	root       string
	projectID  string
	jsonOutput bool
	client     *http.Client
	stdin      io.Reader
	stdout     io.Writer
	openURL    browserOpener
}

func run(args []string) error {
	return runWithIO(args, os.Stdin, os.Stdout, openBrowser)
}

func runWithIO(args []string, stdin io.Reader, stdout io.Writer, openURL browserOpener) error {
	fs := flag.NewFlagSet("prilog", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	apiURLFlag := fs.String("api-url", "", "Prilog API URL")
	jsonFlag := fs.Bool("json", false, "Return machine-readable JSON without interactive authentication")
	projectFlag := fs.String("project-id", "", "Explicit existing project UUID")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printUsage(stdout)
			return nil
		}
		return err
	}

	args = fs.Args()
	if len(args) == 0 {
		printUsage(stdout)
		return nil
	}

	root, err := repoRoot()
	if err != nil {
		return err
	}

	c := newCLI(configuredAPIURL(*apiURLFlag), root, stdin, stdout, openURL)
	c.jsonOutput = *jsonFlag
	c.projectID = strings.TrimSpace(*projectFlag)
	if c.projectID != "" && !validProjectID(c.projectID) {
		return errors.New("--project-id must be a UUID")
	}
	if c.jsonOutput {
		switch args[0] {
		case "status", "projects", "create-project", "use-project", "add-repo", "list", "ingest", "issue", "fix", "diff", "pr":
		default:
			return errors.New("--json supports status, projects, create-project, use-project, add-repo, list, ingest, issue import, fix, diff, and pr")
		}
	}
	return c.dispatch(args)
}

func newCLI(apiURL, root string, stdin io.Reader, stdout io.Writer, openURL browserOpener) *cli {
	if stdin == nil {
		stdin = os.Stdin
	}
	if stdout == nil {
		stdout = os.Stdout
	}
	if openURL == nil {
		openURL = func(string) error { return nil }
	}
	return &cli{
		apiURL:  strings.TrimRight(apiURL, "/"),
		root:    root,
		client:  &http.Client{Timeout: 60 * time.Second},
		stdin:   stdin,
		stdout:  stdout,
		openURL: openURL,
	}
}

func (c *cli) dispatch(args []string) error {
	command := args[0]
	rest := args[1:]
	if handled, err := dispatchCLIExtension(c, command, rest); handled {
		return err
	}

	switch command {
	case "help", "-h", "--help":
		if len(rest) > 0 {
			return errors.New("help does not accept arguments")
		}
		printUsage(c.stdout)
		return nil
	case "version", "-v", "--version":
		if len(rest) > 0 {
			return errors.New("version does not accept arguments")
		}
		printVersion(c.stdout)
		return nil
	case "login":
		if len(rest) > 0 {
			return errors.New("login does not accept arguments")
		}
		return c.login(true)
	case "logout":
		if len(rest) > 0 {
			return errors.New("logout does not accept arguments")
		}
		return c.logout()
	case "projects":
		if len(rest) > 0 {
			return errors.New("projects does not accept arguments")
		}
		return c.projects()
	case "create-project":
		return c.createProject(rest)
	case "use-project":
		return c.useProject(rest)
	case "add-repo":
		if len(rest) > 0 {
			return errors.New("add-repo does not accept arguments")
		}
		return c.addRepo()
	case "config":
		return c.config(rest)
	case "status":
		if len(rest) > 0 {
			return errors.New("status does not accept arguments")
		}
		return c.status()
	case "ingest":
		if len(rest) > 1 {
			return errors.New("ingest accepts at most one file")
		}
		return c.ingest(rest)
	case "issue":
		return c.issue(rest)
	case "list":
		if len(rest) > 1 {
			return errors.New("list accepts at most one filter")
		}
		status := ""
		if len(rest) == 1 {
			status = rest[0]
		}
		return c.list(status)
	case "fix":
		id, err := singleIDArg("fix", rest)
		if err != nil {
			return err
		}
		return c.fix(id)
	case "diff":
		id, err := singleIDArg("diff", rest)
		if err != nil {
			return err
		}
		return c.diff(id)
	case "pr":
		id, err := singleIDArg("pr", rest)
		if err != nil {
			return err
		}
		return c.pr(id)
	default:
		return fmt.Errorf("unknown command %q", command)
	}
}

func singleIDArg(command string, args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("%s requires an error id", command)
	}
	if len(args) > 1 {
		return "", fmt.Errorf("%s accepts exactly one error id", command)
	}
	return cleanID(args[0]), nil
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `Usage:
  prilog [--json] [--project-id <uuid>] <command> (flags precede the command)
  prilog login
  prilog logout
  prilog projects
  prilog create-project <project-name>
  prilog use-project <project-name|id>
  prilog add-repo
  prilog config
  prilog status
  prilog ingest [file]
  prilog issue import <github-issue-url|owner/repo#number>
  prilog list [pending|processing|completed|failed]
  prilog fix <id>
  prilog diff <id>
  prilog pr <id>
  prilog version`)
	printCLIExtensionUsage(w)
}

func printVersion(w io.Writer) {
	fmt.Fprintf(w, "prilog %s\n", version)
}

func (c *cli) println(values ...any) {
	fmt.Fprintln(c.stdout, values...)
}

func (c *cli) printf(format string, values ...any) {
	fmt.Fprintf(c.stdout, format, values...)
}
