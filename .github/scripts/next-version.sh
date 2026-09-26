#!/usr/bin/env bash
# Uso: next-version.sh [auto|major|minor|patch]  -> imprime a proxima versao (sem "v")
set -euo pipefail

force="${1:-auto}"
last_tag=$(git describe --tags --abbrev=0 --match 'v[0-9]*' 2>/dev/null || true)

if [ -n "$last_tag" ]; then
  current="${last_tag#v}"
  range="${last_tag}..HEAD"
else
  current="$(tr -d '[:space:]' < VERSION)"
  range="HEAD"
fi

if [ "$(git rev-list --count "$range")" -eq 0 ]; then
  echo "Nenhum commit desde ${last_tag:-o inicio}" >&2
  exit 1
fi

subjects=$(git log --format=%s "$range")
bodies=$(git log --format=%b "$range")

bump=patch
if echo "$subjects" | grep -Eq '^feat(\(.+\))?:'; then bump=minor; fi
if echo "$subjects" | grep -Eq '^[a-z]+(\(.+\))?!:' || echo "$bodies" | grep -q '^BREAKING CHANGE'; then bump=major; fi
if [ "$force" != auto ]; then bump="$force"; fi

IFS=. read -r major minor patch <<< "$current"
case "$bump" in
  major) echo "$((major + 1)).0.0" ;;
  minor) echo "${major}.$((minor + 1)).0" ;;
  patch) echo "${major}.${minor}.$((patch + 1))" ;;
  *) echo "bump invalido: $bump" >&2; exit 1 ;;
esac
