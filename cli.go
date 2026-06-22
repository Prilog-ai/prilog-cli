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
	apiURL  string
	root    string
	client  *http.Client
	stdin   io.Reader
	stdout  io.Writer
	openURL browserOpener
}

func run(args []string) error {
	return runWithIO(args, os.Stdin, os.Stdout, openBrowser)
}

func runWithIO(args []string, stdin io.Reader, stdout io.Writer, openURL browserOpener) error {
	fs := flag.NewFlagSet("prilog", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	apiURLFlag := fs.String("api-url", "", "Prilog API URL")
	if err := fs.Parse(args); err != nil {
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
	case "init":
		if len(rest) > 0 {
			return errors.New("init does not accept arguments")
		}
		return c.init()
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
  prilog login
  prilog init
  prilog config
  prilog status
  prilog ingest [file]
  prilog list [pending|processing|completed|failed]
  prilog fix <id>
  prilog diff <id>
  prilog pr <id>
  prilog version`)
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
