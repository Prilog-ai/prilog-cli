# Prilog CLI

`prilog` brings Prilog into your terminal: connect a repository, ship logs or traces, review generated fixes, and open pull requests from the command line.

## Install

macOS and Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/prilog-ai/prilog-cli/main/install.sh | sh
```

The installer detects your operating system and CPU architecture, downloads the matching release asset, verifies checksums when available, and installs the `prilog` binary into `/usr/local/bin`.

No API setup is required after installation; the CLI talks to the production Prilog API by default.

To install somewhere else:

```sh
curl -fsSL https://raw.githubusercontent.com/prilog-ai/prilog-cli/main/install.sh | PRILOG_INSTALL_DIR="$HOME/.local/bin" sh
```

## Quick Start

Authenticate once:

```sh
prilog login
```

Connect the current repository to a Prilog project:

```sh
prilog init
```

`init` uses the repository name as the default project name, lets you rename it, writes `.prilog/config.json`, and opens the Prilog onboarding flow for that project.

Check the connected account and project:

```sh
prilog status
```

Send a log, trace, or error file:

```sh
prilog ingest ./logs.log
```

Work with detected errors:

```sh
prilog list pending
prilog fix <error-id>
prilog diff <error-id>
prilog pr <error-id>
```

## Commands

| Command | Description |
| --- | --- |
| `prilog login` | Authenticate or switch Prilog accounts. |
| `prilog init` | Link the current repository to a Prilog project and launch onboarding. |
| `prilog status` | Show the active user, organization, project, log totals, and fix totals. |
| `prilog config` | Show local CLI configuration for the current repository. |
| `prilog ingest [file]` | Upload logs, traces, or errors from a file or stdin. |
| `prilog list [filter]` | List recent logs and errors. Filters: `all`, `pending`, `processing`, `completed`, `failed`. |
| `prilog fix <id>` | Queue Prilog analysis for an error. |
| `prilog diff <id>` | Print the generated fix diff in the terminal. |
| `prilog pr <id>` | Create a pull request for the generated fix and print the PR URL. |
| `prilog version` | Print the installed CLI version. |

## Security

Authentication uses a browser-based Prilog flow. Access and refresh tokens are stored in your operating system's user config directory, while repository project metadata is stored in `.prilog/config.json`.

Every command that reads or mutates Prilog data is authenticated against the Prilog API and authorized against the active organization and project.

## Development

Requirements:

- Go 1.23 or newer
- Git

Common tasks:

```sh
go test ./...
go vet ./...
go build -o prilog .
```

The CLI is intentionally stdlib-only. The source is split around user-facing commands, auth, HTTP transport, repository metadata, config persistence, and output formatting so the command surface stays easy to audit.
