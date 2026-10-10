#!/usr/bin/env bash
# Seed a supplied existing review project; no login, project creation or PRs.
set -euo pipefail
if [[ $# != 2 || "$1" != "--project-id" ]]; then
  printf 'Usage: scripts/review-plugin.sh --project-id <existing-review-project-uuid>\n' >&2
  exit 2
fi
plugin_root=$(cd "$(dirname "$0")/.." && pwd)
prilog_binary=${PRILOG_BINARY:-prilog}
"$prilog_binary" --json --project-id "$2" ingest "$plugin_root/review-fixtures/payments/error.log"
"$prilog_binary" --json --project-id "$2" list all
