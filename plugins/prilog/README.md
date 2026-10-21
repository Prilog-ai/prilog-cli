# Prilog plugin

Inspect production errors, request fixes, review diffs, and create pull requests
in the Prilog projects you authorize.

The hosted MCP server is `https://prilog.ai/api/mcp`. Connect through Prilog's
OAuth consent flow and select the organization, projects, and permissions to
share. Revoke access in Prilog Settings → Account → Connected apps.

In environments with command execution, the shared skill uses the separately
installed Prilog CLI. See [CLI installation and usage](https://github.com/Prilog-ai/prilog-cli).
The CLI manages its own browser login; the skill does not read credential files.

This generated directory contains only the plugin manifest, hosted MCP
configuration, shared skill, static branding images, this README, and license.
The images are listing assets, not executable programs. CLI implementation,
installers, release tooling, and test fixtures remain outside this directory.
Maintain the shared sources in the repository root and regenerate with
`python scripts/package-plugin.py`.

[Privacy policy](https://prilog.ai/privacy-policy/) ·
[Terms of service](https://prilog.ai/terms-of-service/) ·
[Support](https://prilog.ai)
