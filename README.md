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

Connect the current repository to a Prilog project:

```sh
prilog init
```

Check the active account, organization, project, and fix totals:

```sh
prilog status
```

Ingest a local log, trace, or error file:

```sh
prilog ingest ./logs.log
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
| `prilog init` | Link the current repository to a Prilog project and launch onboarding. |
| `prilog status` | Show the active user, organization, project, log totals, and fix totals. |
| `prilog config` | Show local CLI configuration for the current repository. |
| `prilog config path` | Print global and repository config file paths. |
| `prilog ingest [file]` | Upload logs, traces, or errors from a file or stdin. |
| `prilog list [filter]` | List recent logs and errors. Filters: `all`, `pending`, `processing`, `completed`, `failed`. |
| `prilog fix <id>` | Queue Prilog analysis for an error. |
| `prilog diff <id>` | Print the generated fix diff in the terminal. |
| `prilog pr <id>` | Create a pull request for the generated fix and print the PR URL. |
| `prilog version` | Print the installed CLI version. |

## 🧪 Example Workflow

```sh
# Authenticate and connect this repository
prilog login
prilog init

# Send telemetry from a local file
prilog ingest ./production-errors.log

# Find pending issues and start a fix
prilog list pending
prilog fix 018f4a2e-7c2b-7b9d-ae4a-0ef5d4f9a101

# Review locally or create a PR
prilog diff 018f4a2e-7c2b-7b9d-ae4a-0ef5d4f9a101
prilog pr 018f4a2e-7c2b-7b9d-ae4a-0ef5d4f9a101
```

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
