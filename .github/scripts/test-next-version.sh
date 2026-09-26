#!/usr/bin/env bash
set -euo pipefail
SCRIPT="$(cd "$(dirname "$0")" && pwd)/next-version.sh"
fail=0

new_repo() {
  d=$(mktemp -d); cd "$d"
  git init -q -b develop
  git config user.email t@t; git config user.name t
  echo 0.0.0 > VERSION; git add VERSION; git commit -qm "chore: init"
}
commit() { if [ -n "${2:-}" ]; then git commit -q --allow-empty -m "$1" -m "$2"; else git commit -q --allow-empty -m "$1"; fi; }
check() { # expected actual label
  if [ "$1" != "$2" ]; then echo "FAIL $3: expected '$1' got '$2'"; fail=1; else echo "ok   $3"; fi
}

new_repo; commit "feat: a"
check 0.1.0 "$("$SCRIPT")" "sem tag, feat -> 0.1.0"

new_repo; commit "docs: a"
check 0.0.1 "$("$SCRIPT")" "sem tag, docs -> patch"

new_repo; git tag v1.2.3; commit "fix: x"
check 1.2.4 "$("$SCRIPT")" "fix -> patch"

commit "feat(api): y"
check 1.3.0 "$("$SCRIPT")" "feat -> minor"

commit "feat!: z"
check 2.0.0 "$("$SCRIPT")" "feat! -> major"

new_repo; git tag v1.2.3; commit "fix: x" "BREAKING CHANGE: mudou contrato"
check 2.0.0 "$("$SCRIPT")" "BREAKING CHANGE no corpo -> major"

new_repo; git tag v1.2.3; commit "feat: y"
check 1.2.4 "$("$SCRIPT" patch)" "bump forcado patch"
check 2.0.0 "$("$SCRIPT" major)" "bump forcado major"

new_repo; git tag v1.2.3
if "$SCRIPT" >/dev/null 2>&1; then echo "FAIL sem commits deveria falhar"; fail=1; else echo "ok   sem commits falha"; fi

exit $fail
