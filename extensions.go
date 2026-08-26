package main

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type cliExtension struct {
	usage string
	run   func(*cli, []string) error
}

var cliExtensions = map[string]cliExtension{}

func registerCLIExtension(name, usage string, run func(*cli, []string) error) {
	name = strings.TrimSpace(name)
	if name == "" || run == nil {
		panic("invalid CLI extension")
	}
	if _, exists := cliExtensions[name]; exists {
		panic(fmt.Sprintf("CLI extension %q is already registered", name))
	}
	cliExtensions[name] = cliExtension{usage: strings.TrimSpace(usage), run: run}
}

func dispatchCLIExtension(c *cli, command string, args []string) (bool, error) {
	extension, ok := cliExtensions[command]
	if !ok {
		return false, nil
	}
	return true, extension.run(c, args)
}

func printCLIExtensionUsage(w io.Writer) {
	if len(cliExtensions) == 0 {
		return
	}

	names := make([]string, 0, len(cliExtensions))
	for name := range cliExtensions {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Fprintln(w, "\nAdditional tools:")
	for _, name := range names {
		usage := cliExtensions[name].usage
		if usage == "" {
			usage = "prilog " + name
		}
		fmt.Fprintf(w, "  %s\n", usage)
	}
}
