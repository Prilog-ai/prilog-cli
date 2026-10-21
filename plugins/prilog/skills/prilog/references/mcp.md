# Hosted MCP workflows

Use this reference only for an explicitly requested hosted connector or an environment without CLI command execution. In command-enabled environments, use `prilog --help` and the CLI workflow. Do not switch to MCP to work around a CLI authentication or project-setup error.

Connect the Prilog plugin using account authorization. Consent grants one organization, selected projects, and requested capabilities. To add projects or permissions, reconnect from the chat app. Access can be revoked under Prilog Settings → Account → Connected apps.

Tool results contain `http_status` and `data`. Error lists return summaries; messages longer than 1,000 characters set `message_truncated`. Use `get_error` for the full error and analysis. Start with `list_projects`, then pass `project_id` on every other tool. If more than one project matches the request, resolve the choice with the user. Do not invent identifiers.

| Tool | Required inputs | Optional inputs |
| --- | --- | --- |
| list_projects | none | none |
| get_status | project_id | none |
| list_errors | project_id | status, search, page, page_size (1–100) |
| get_error / get_diff | project_id, error_id | none |
| ingest_logs | project_id, text, idempotency_key | filename |
| import_github_issue | project_id, issue_url, idempotency_key | none |
| request_fix / create_pull_request | project_id, error_id, idempotency_key | none |

For each requested mutation choose a unique 8–128 character idempotency key, such as a UUID. Reuse the same key and exact arguments when retrying that action. Never switch keys to evade a running or uncertain outcome. Check the error/project state first; an interrupted mutation may require Prilog support to reconcile it.

Upload only user-supplied text (up to 1 MiB) or explicitly selected local content. The hosted connector cannot read local filesystem paths. Issue import accepts an open public GitHub issue for the project's connected fork, subject to existing Prilog rules.

`request_fix` returns promptly with an error ID and review URL. Use `get_error` to check status, at most six checks spaced about ten seconds apart when the environment supports waiting. Otherwise provide the review link and pending state. On completion, call `get_diff`. Empty diffs can mean analysis is pending, failed, skipped, or produced no code changes; inspect the status before proceeding.

Create a PR only when the user requested publication and the generated fix is available. Call `get_diff` first. Existing publications are returned through the same workflow; report the PR URL in `data.pr_url`, `data.resolution_url`, or resolution metadata. Never merge or deploy through this plugin.
