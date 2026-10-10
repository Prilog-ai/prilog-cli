# CLI workflows

Run `prilog --help` first. The installed CLI's help is the source of truth for available commands and flags. Execute Prilog operations through that binary and parse its output as needed. Do not discover its commands or state by reading configuration, inspecting implementation files, or searching the repository.

For account, active project, available projects, and pending errors, use:

```sh
prilog --help
prilog --json status
prilog --json projects
prilog --json --project-id <project-id-from-status> list pending
```

Use `--json` and `--project-id` only when the installed CLI supports them. With JSON output, obtain the project ID from `project.id` in the status JSON. If the user already supplied a project ID, pass it to `status` too. If no project override is supported, retain the CLI's existing selection. Older CLI versions can use human-readable commands; do not parse tables as JSON. Use `prilog version` when an unsupported command or flag needs diagnosis.

Let the CLI read its own saved account, project, endpoint, and credentials. Use `prilog config` only when configuration information is needed. Never open the underlying files, inspect token stores, or search the product for those values.

If the executable is missing, report that Prilog must be installed. Installation instructions are at https://github.com/Prilog-ai/prilog-cli. If authentication is required, report `prilog login`; if repository setup is required, report `prilog add-repo`. Execute these setup commands when the user requested installation, sign-in, project creation, or repository setup. Do not silently install dependencies or repeatedly launch login flows. Login requires browser interaction. For an explicitly requested logout or signed-out test, use `prilog logout` when available in help; let the CLI clear its saved login.

Use `prilog projects` to discover all projects in the current organization. Use `prilog use-project <name>` when the user selects one, or its ID if the name is ambiguous. `prilog create-project <name>` creates and switches to the new project; names are required in noninteractive automation. These commands persist a CLI default across working directories. Keep using `--project-id` when a request needs an explicit override.

For first-time setup, use `prilog login`, then `prilog add-repo`. The latter opens Prilog and immediately starts GitHub authorization in the selected project; the browser returns to the existing Add service wizard to finish repository/monitoring onboarding. Wait for the user to finish the browser wizard before ingesting or requesting fixes that need the repository. Run `add-repo` again to connect another repository in the current project. `--json add-repo` returns a browser URL without opening it. If the user wants a named workspace first, use `create-project <name>` before `add-repo`. The old `init` command is removed.

Place global flags before the command. These are examples for the current CLI; use installed help to confirm support:

```sh
prilog --json projects
prilog --json create-project "Payments"
prilog --json use-project "Payments"
prilog add-repo
prilog --json --project-id <project-uuid> status
prilog --json --project-id <project-uuid> list pending
prilog --json --project-id <project-uuid> ingest ./selected-errors.log
prilog --json --project-id <project-uuid> issue import https://github.com/owner/repo/issues/123
prilog --json --project-id <project-uuid> fix <error-uuid>
prilog --json --project-id <project-uuid> diff <error-uuid>
prilog --json --project-id <project-uuid> pr <error-uuid>
```

`--json` returns a structured command response and fails instead of prompting for authentication. Refreshing an existing login remains supported.

`fix` queues asynchronous analysis using the project's configured automation. Check progress with `list processing`, `list completed`, and `diff`, at most six checks spaced about ten seconds apart when waiting is supported. Otherwise report the pending state. Review `diff` before requested `pr` publication. After an interrupted mutation, inspect state through the CLI before retrying; the CLI has no MCP idempotency-key contract.

Honor the CLI's configured API endpoint and any explicit endpoint the user supplied. Do not infer an endpoint from product files or switch a local test to production on failure.
