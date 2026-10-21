# Prilog CLI

[![Latest release](https://img.shields.io/github/v/release/Prilog-ai/prilog-cli?label=release)](https://github.com/Prilog-ai/prilog-cli/releases)
[![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Homebrew](https://img.shields.io/badge/Homebrew-prilog--ai%2Ftap-FBB040?logo=homebrew&logoColor=white)](https://github.com/Prilog-ai/homebrew-tap)
[![License](https://img.shields.io/github/license/Prilog-ai/prilog-cli)](LICENSE)

`prilog` brings Prilog into your terminal: connect a repository, ship logs and traces, review drafted fixes, and open pull requests without leaving your shell.

## ✨ What It Does

- 🔐 Authenticates with Prilog through a browser-based login flow.
- 📁 Links the current Git repository to a Prilog project.
- 📡 Ingests logs, traces, and error files from local files or stdin.
- 🐛 Imports an open issue from a public GitHub upstream for analysis against a connected fork.
- 🧭 Lists detected errors by status, including pending fixes.
- 🛠️ Starts Prilog analysis for an error and opens the dashboard review.
- 🔎 Prints generated diffs directly in the terminal.
- 🚀 Creates pull requests for generated fixes in the connected repository.

## 📦 Install

### Homebrew

```sh
brew install prilog-ai/tap/prilog
```

### One-Line Installer

macOS and Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/Prilog-ai/prilog-cli/main/install.sh | sh
```

The installer detects your operating system and CPU architecture, downloads the matching release asset, verifies checksums when available, and installs the `prilog` binary into `/usr/local/bin`.

To install somewhere else:

```sh
curl -fsSL https://raw.githubusercontent.com/Prilog-ai/prilog-cli/main/install.sh | PRILOG_INSTALL_DIR="$HOME/.local/bin" sh
```

Verify the install:

```sh
prilog version
```

## ⚡ Quick Start

Authenticate once:

```sh
prilog login
```

For a private automated client with an operator-issued service account token,
set `PRILOG_ACCESS_TOKEN` in the CLI process environment. This skips browser
login and refresh without replacing your saved login. The environment token
does not inherit the saved login's default project; use `--project-id` to select
one explicitly. Service tokens retain the user's workspace permissions and can
be revoked by an operator. Keep them out of browser apps and distributed clients.

Connect the current repository to a Prilog project:

```sh
prilog add-repo
```

Check the active account, organization, project, and fix totals:

```sh
prilog status
```

Ingest a local log, trace, or error file:

```sh
prilog ingest ./logs.log
```

Or import a public GitHub issue after connecting your fork of that repository:

```sh
prilog issue import https://github.com/medusajs/medusa/issues/16487
```

Review and act on detected errors:

```sh
prilog list pending
prilog fix <error-id>
prilog diff <error-id>
prilog pr <error-id>
```

## 🧰 Commands

| Command | Description |
| --- | --- |
| `prilog login` | Authenticate or switch Prilog accounts. |
| `prilog logout` | Clear the saved CLI login while keeping endpoint and repository settings. |
| `prilog projects` | List all projects in your current organization and mark the selected default. |
| `prilog create-project <project-name>` | Create a project and select it as the CLI default. |
| `prilog use-project <project-name\|id>` | Select an existing default project. |
| `prilog add-repo` | Open Prilog, authorize GitHub, and continue Add service onboarding in the selected project. |
| `prilog status` | Show the active user, organization, project, log totals, and fix totals. |
| `prilog config` | Show local CLI configuration for the current repository. |
| `prilog config path` | Print global and repository config file paths. |
| `prilog ingest [file]` | Upload logs, traces, or errors from a file or stdin. |
| `prilog issue import <url\|owner/repo#number>` | Import an open public GitHub issue into the active project. The connected code repository must be your fork of that upstream. |
| `prilog list [filter]` | List recent logs and errors. Filters: `all`, `pending`, `processing`, `completed`, `failed`, `skipped`. Skipped items were stopped before analysis. |
| `prilog fix <id>` | Queue Prilog analysis for an error. |
| `prilog diff <id>` | Print the generated fix diff in the terminal. |
| `prilog pr <id>` | Create a pull request for the generated fix and print the PR URL. |
| `prilog version` | Print the installed CLI version. |

## 🧪 Example Workflow

```sh
# Authenticate and connect the first GitHub repository
prilog login
prilog add-repo

# Send telemetry from a local file
prilog ingest ./production-errors.log

# Find pending issues and start a fix
prilog list pending
prilog fix 018f4a2e-7c2b-7b9d-ae4a-0ef5d4f9a101

# Review locally or create a PR
prilog diff 018f4a2e-7c2b-7b9d-ae4a-0ef5d4f9a101
prilog pr 018f4a2e-7c2b-7b9d-ae4a-0ef5d4f9a101
```

### Public upstream contribution

Use one Prilog project per upstream repository. Fork the upstream into the GitHub account connected to Prilog, install the Prilog GitHub App on that fork, and select the upstream default branch during service setup. Then:

```sh
prilog issue import medusajs/medusa#16487
prilog fix <error-id>
prilog diff <error-id>
prilog pr <error-id>
```

The importer rejects closed issues, pull request URLs, unrelated forks, and forks configured on the wrong base branch. `prilog pr` pushes the generated branch to the connected fork and opens a cross-fork pull request against the issue's upstream repository with the connected GitHub user as the PR author.

## 🔐 Security

Authentication uses a browser-based Prilog flow. Access and refresh tokens are stored in your operating system's user config directory, while repository project metadata is stored in `.prilog/config.json`.

Every command that reads or mutates Prilog data is authenticated against the Prilog API and authorized against the active organization and project.

Recommended hygiene:

- Do not commit `.prilog/config.json` if your repository policy treats project metadata as private.
- Run `prilog login` again to switch accounts or refresh local credentials.
- Use `prilog config path` to find and remove local config files when testing a clean setup.

## 🧑‍💻 Development

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

## 🤝 Support

- Product and CLI issues: [Prilog-ai/prilog-cli issues](https://github.com/Prilog-ai/prilog-cli/issues)
- Homebrew packaging issues: [Prilog-ai/homebrew-tap issues](https://github.com/Prilog-ai/homebrew-tap/issues)
- Website: [prilog.ai](https://prilog.ai)

## Agent plugin and automation

The shared Prilog skill and hosted MCP configuration ship in this repository.
The hosted connector operates on existing projects you authorize. Configure it
in ChatGPT or Claude using `https://prilog.ai/api/mcp` and the Prilog account
consent flow; revoke access in Prilog Settings → Account → Connected apps.

For local automation, put `--json` and `--project-id <uuid>` before the command.
`--json` supports status, projects, create-project, use-project, add-repo, list, ingest, issue import, fix, diff, and pr. It returns
a structured command response without interactive authentication; run `prilog login` first.
`PRILOG_CONFIG_DIR` can isolate login configuration in an absolute directory.

Build and validate the release archives in a Python environment with
`jsonschema==4.26.0` installed:

```sh
python scripts/package-plugin.py
python scripts/package-plugin.py --check
```

The generator uses root `plugin.json`, `mcp.json`, `skills/`, and `assets/` as
shared sources. It maintains `.claude-plugin/plugin.json` and `.mcp.json` for
Claude repository installation and writes both ZIPs under `dist/plugin/`.
Only plugin resources and the license enter the archives. The plugin has its
own version in `plugin.json`; update it when releasing skill or package changes.
For Claude directory submission, use repository `Prilog-ai/prilog-cli`, branch
`main`, and subfolder `plugins/prilog`. This generated folder contains the Claude
package and a plugin README, without the CLI source or installer. The validator
checks that it matches the shared sources and contains no unexpected files.
See `PLUGIN_PUBLISHING.md` for deployment, review fixtures, and store submission.

## Project selection and repository setup

`prilog create-project "Payments"` creates and selects a project. Run
`prilog projects` to see your organization's projects and `prilog use-project
"Payments"` to switch the CLI default. Selection persists across directories;
`--project-id <uuid>` overrides it for one command. Names match exactly without
case sensitivity. If a name is duplicated, select its ID from `prilog projects`.
An interactive `create-project` with no name prompts for one; automation must
supply the name.

`prilog add-repo` uses the selected project and opens a browser handoff to Prilog,
which redirects to GitHub authorization and returns to the existing Add service
repository wizard. Run it again to add another repository/service in the same
project. The browser must sign in as the CLI's Prilog account. If no project is
selected, the first Add service flow uses the account's current/default project.
`--json add-repo` returns the handoff URL without opening a browser.

The `init` command has been removed. Project/repository setup now uses
`create-project`, `use-project`, and `add-repo`. The new API and web changes must
be deployed before using `add-repo` against a hosted Prilog instance.
