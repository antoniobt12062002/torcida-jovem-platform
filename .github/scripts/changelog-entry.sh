#!/usr/bin/env bash
# Uso: changelog-entry.sh VERSION -> imprime a secao do CHANGELOG (commits desde a ultima tag)
set -euo pipefail

version="$1"
last_tag=$(git describe --tags --abbrev=0 --match 'v[0-9]*' 2>/dev/null || true)
range="${last_tag:+${last_tag}..}HEAD"

echo "## v${version} - $(date -u +%Y-%m-%d)"
echo

section() { # titulo regex
  lines=$(git log --no-merges --format='- %s (%h)' "$range" | grep -E "^- $2" || true)
  if [ -n "$lines" ]; then echo "### $1"; echo "$lines"; echo; fi
}

section "Funcionalidades" 'feat(\(.+\))?!?:'
section "Correcoes" 'fix(\(.+\))?!?:'
section "Outras mudancas" '(docs|chore|refactor|perf|test|ci|build|style)(\(.+\))?!?:'
