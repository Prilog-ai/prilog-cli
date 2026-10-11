# Prilog plugin release and review

The plugin is implemented but stays behind `MCP_ENABLED=false` until deployment.
The archives are submission inputs, not evidence of store approval. The server
and account authorization must be publicly reachable before either submission.

## Deploy the connector

1. Deploy application migration 102 for connection grants and durable mutation results.
2. Set `MCP_ENABLED=true` and the same `PRILOG_MCP_RESOURCE=https://prilog.ai/api/mcp`
   in the API and identity service. Run the existing identity migration command
   (`npm run db:migrate` in `auth-service`) with MCP enabled to create the OAuth
   and signing-key tables. Keep Go IDP configuration enabled.
3. Deploy the API, identity service, and translated web UI together. The platform
   Terraform `mcp_enabled` variable defaults to false and configures both services.
   V1 uses resource-bound bearer tokens and does not advertise DPoP.
   Identity routing includes OAuth discovery; internal verification remains private.
4. Verify `/api/mcp`, the root and path-inserted protected-resource discovery URLs,
   and `https://auth.prilog.ai/.well-known/oauth-authorization-server`. The MCP URL
   must be the token audience and must not be served from a redirect.
5. Connect through each platform's custom-connector flow before directory submission.
   Test login from a signed-out browser, a selected existing project, viewer denial
   of writes, revocation under Settings → Account → Connected apps, and reconnect.

Application and identity migrations are additive. Roll back service exposure by
setting `MCP_ENABLED=false`; retain grant and mutation records during rollback.
Do not drop mutation records while any action is running or has an uncertain outcome.

For a local environment, use `PRILOG_MCP_RESOURCE=http://localhost:3333/api/mcp`
and the local IDP/API/UI URLs. Do not send local smoke tests to production.

## Prepare the review workspace

Create a dedicated reviewer account and organization with two existing projects,
only one selected during connector consent. Supply account access to reviewers in
private portal fields, never in this repository or plugin ZIP.

Use `review-fixtures/payments/` as the source for a dedicated review repository.
Connect its fork through the existing GitHub onboarding flow and install the
Prilog GitHub App on that fork. Point the review project at the correct branch.
The supplied `error.log` corresponds to a reproducible divide-by-zero failure in
`main.go`. Seed it by running `scripts/review-plugin.sh --project-id <uuid>` with
an already authenticated CLI. The script only ingests the fixture and lists errors.
Use approval-required project automation so fixture ingestion or fix generation
does not automatically publish a PR. Keep a completed analysis available for reviewers so diff/PR cases do not depend
on analysis completing during their first session.

Use these review prompts in both products:

| Prompt | Expected result |
| --- | --- |
| Show my authorized projects and pending errors. | Only the selected project; stable error IDs and observed status. |
| Upload this log to my review project. | Accepted/skipped counts; repeating the same MCP key does not duplicate the action. |
| Generate a fix for this error and show me the diff. | Queued state first, then generated changes or a truthful pending/failed state. |
| Create a PR for this reviewed fix. | A new or existing PR URL on the connected fork; no merge or deployment. |
| Explain this error without changing anything. | Read-only tools; no fix, ingestion, or publication. |
| Show errors from the project I did not authorize. | Permission denial; no data from the other project. |

Also test issue import with an open public issue on the upstream of the connected
fork. Existing repository and competing-PR rules still apply. No credentials,
customer logs, or live private repository data belong in review fixtures.

## Repeatable local agent sessions

Run `python3 scripts/test-agent-sessions.py` from the CLI repository with authenticated
Claude Code and Codex CLIs available. Each case uses a new temporary session, the real
Prilog binary, synthetic credentials, a loopback fixture API, and recorded browser
handoffs. It leaves your saved Prilog login and production data unchanged.

Use `--prilog-binary /absolute/path/to/prilog` to test a build and `--output /private/tmp/prilog-agent-tests`
to choose the report directory. `results.json` includes command traces, requests,
observed responses, and failures. These cases test agent/CLI behavior; they do not
replace real OAuth, GitHub onboarding, analysis, or PR publication on a review workspace.

## Manually test the local plugin

From the monorepo root, build the new CLI and load the Claude plugin directory:

```sh
(cd cli && go build -o /private/tmp/prilog .)
claude plugin validate ./cli
PATH="/private/tmp:$PATH" claude --plugin-dir ./cli
```

Invoke `/prilog:prilog` and explicitly ask for the local CLI workflow on an
existing test project. If needed, authenticate using `prilog login` outside JSON
mode first. The packaged MCP URL targets the hosted service, so hosted tests
require deployment even though the local skill can be loaded immediately.

After deployment, add `https://prilog.ai/api/mcp` as a custom OAuth-authenticated
connector in both ChatGPT and Claude. Select one test project, run the review
prompts above, and check that the unselected project is denied. Revoke the
connection under Prilog Settings → Account → Connected apps and verify that
subsequent tool calls require reconnecting.

## Build the packages

Use Python with `jsonschema==4.26.0` installed:

```sh
python scripts/package-plugin.py
python scripts/package-plugin.py --check
```

Validate the skill with an Agent Skills frontmatter validator. CI validates the
portable manifest and MCP schemas, generated Claude manifests, packaged resource
paths, and allowlisted archive contents. ZIPs live under `dist/plugin/`; the CLI
release workflow attaches them to the release after its binary release step.

Shared content is maintained once. The OpenAI ZIP contains root `plugin.json`
and `mcp.json`. The Claude ZIP contains `.claude-plugin/plugin.json` and
`.mcp.json`; these generated files also make the CLI GitHub repository installable.
Do not submit app-reference files or lifecycle hooks. Update `plugin.json`'s
version when changing skills, metadata, or package resources, then regenerate.

## Submit and monitor

OpenAI portal: https://platform.openai.com/plugins

OpenAI: verify the publishing identity, upload the OpenAI ZIP with its MCP
configuration, complete domain verification and OAuth connection, resolve tool
scan findings, supply private review account access, and submit for review.

Claude portal: https://claude.ai/directory/manage/new

Claude: use the directory submission portal and submit the GitHub repository
containing the generated Claude plugin and shared skill. Configure the same
remote endpoint, supply the review workspace, and resolve validation findings.
If either portal requires listing screenshots, capture the deployed review flow;
do not submit fabricated screenshots or claim unrun reviewer scenarios passed.

Once approved, publish from each portal. Monitor `mcp tool=… status=… duration_ms=…`
logs, connection/auth failures, rate-limit responses, and unresolved mutation rows.
The mutation journal records connection, request hash, resource, outcome, and
completion time without storing ingestion text or access tokens.

An interrupted action remains `running` and blocks conflicting writes. An
operator must inspect the actual error/analysis/PR state and reconcile the stored
result before clearing that block. Never delete the row just to retry a PR.

Official guidance checked during implementation:
- https://developers.openai.com/plugins/build/plugins
- https://developers.openai.com/plugins/deploy/submission
- https://developers.openai.com/plugins/build/auth
- https://claude.com/resources/articles/build-plugins-for-claude
- https://better-auth.com/docs/plugins/mcp

## CLI project and Add service changes

The local CLI now exposes `projects`, `create-project`, `use-project`, and
`add-repo`; `init` is removed. A selected CLI project is saved with the login and
used across directories. Each request still sends an explicit project header.
Logout/account changes clear the saved account-bound selection.

Deploy the updated API and browser route together. `POST /api/cli/repos/start`
issues an expiring, one-use browser handoff bound to the CLI user, organization,
and project. `/cli/add-repo` preserves it through browser sign-in, then calls
`POST /api/cli/repos/continue` to begin GitHub authorization. A different browser
account, expired link, lost admin access, or blocked billing cannot start setup.
The GitHub callback returns to the existing service-connection wizard; it does
not configure or replace a repository automatically.
