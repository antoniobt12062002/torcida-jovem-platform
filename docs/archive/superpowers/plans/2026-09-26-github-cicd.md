# GitHub Branch Flow, Versioning and CI/CD Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Configure `torcida-jovem-platform` with protected `main`/`develop`, `release/vX.Y.Z` branches deployed to staging, semver versioning and GitHub Actions pipelines.

**Architecture:** Bash scripts compute the next semver and changelog from Conventional Commits. Five workflows (`ci`, `cut-release`, `deploy-staging`, `release`) plus rulesets applied via `gh api`. API/web jobs and deploy steps are placeholders until the stack and hosting exist.

**Tech Stack:** GitHub Actions, GitHub rulesets/Environments, Bash, `gh` CLI, markdownlint-cli2, lychee, dorny/paths-filter, amannn/action-semantic-pull-request.

**Spec:** `docs/superpowers/specs/2026-09-26-github-cicd-design.md`

**Deviations from spec (decided while planning, reflected in Task 12):**
- `pr-title.yml` is folded into `ci.yml` as a job that feeds `ci-gate`. PRs opened by `GITHUB_TOKEN` (the back-merge PR) do not trigger `pull_request` workflows, so a separate required title check would hang forever. `ci.yml` gets a `workflow_dispatch` trigger and `release.yml` dispatches it on the back-merge branch.
- `release/*` ruleset starts with only "no deletion, no force push". Requiring PRs there would block `cut-release` (GITHUB_TOKEN) from pushing the new branch. Tighten after Task 11 validates behavior.
- Back-merge PRs (`main` -> `develop`) MUST use "Create a merge commit" (never squash), otherwise the tag stops being reachable from `develop` and `next-version.sh` miscomputes.

**Working rules:** Tasks 1-9 commit directly to `main` (no rulesets yet; `ci-gate` must exist before it can be required). Confirm with the user before Task 1 (makes repo public) and before Task 11 (creates a real tag).

---

## Chunk 1: Scripts and supporting files

### Task 1: Make repo public and prepare `main`

**Files:** none (GitHub settings)

- [ ] **Step 1: Confirm with the user** that the repo (including PR #1 docs) becomes public.

- [ ] **Step 2: Make the repo public**

```bash
gh repo edit antoniobt12062002/torcida-jovem-platform --visibility public --accept-visibility-change-consequences
gh repo view antoniobt12062002/torcida-jovem-platform --json visibility --jq .visibility
```
Expected: `PUBLIC`

- [ ] **Step 3: Switch to main and confirm clean tree**

```bash
cd "C:/Users/Michels/Desktop/torcida-jovem" && git checkout main && git pull && git status --short
```
Expected: no output (clean).

### Task 2: `next-version.sh` (TDD)

**Files:**
- Create: `.github/scripts/test-next-version.sh`
- Create: `.github/scripts/next-version.sh`

- [ ] **Step 1: Write the failing test**

`.github/scripts/test-next-version.sh`:

```bash
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bash .github/scripts/test-next-version.sh`
Expected: FAIL / error `next-version.sh: No such file or directory`.

- [ ] **Step 3: Write minimal implementation**

`.github/scripts/next-version.sh`:

```bash
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `chmod +x .github/scripts/*.sh && bash .github/scripts/test-next-version.sh`
Expected: 9 lines starting with `ok`, exit code 0.

- [ ] **Step 5: Commit**

```bash
git add .github/scripts
git commit -m "ci: adiciona script de calculo de versao semver

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>"
```

### Task 3: `changelog-entry.sh`

**Files:**
- Create: `.github/scripts/changelog-entry.sh`

- [ ] **Step 1: Write the script**

```bash
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
```

- [ ] **Step 2: Run it against the current history**

Run: `bash .github/scripts/changelog-entry.sh 0.1.0`
Expected: header `## v0.1.0 - <date>` and an "Outras mudancas" list containing the `chore:`/`ci:` commits.

- [ ] **Step 3: Commit**

```bash
git add .github/scripts/changelog-entry.sh
git commit -m "ci: adiciona gerador de entrada do changelog

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>"
```

### Task 4: Version files, lint config, templates, CODEOWNERS, Dependabot

**Files:**
- Create: `VERSION`, `CHANGELOG.md`, `.markdownlint.json`, `.github/CODEOWNERS`, `.github/pull_request_template.md`, `.github/dependabot.yml`

- [ ] **Step 1: Create the files**

`VERSION` (single line):
```
0.0.0
```

`CHANGELOG.md`:
```markdown
# Changelog

```

`.markdownlint.json` (relaxed: existing docs use several H1s and dense lists):
```json
{
  "default": true,
  "MD004": false, "MD007": false, "MD009": false, "MD010": false, "MD012": false,
  "MD013": false, "MD022": false, "MD024": false, "MD025": false, "MD026": false,
  "MD029": false, "MD030": false, "MD031": false, "MD032": false, "MD033": false,
  "MD034": false, "MD036": false, "MD040": false, "MD041": false, "MD046": false,
  "MD047": false
}
```

`.github/CODEOWNERS`:
```
# Adicione o @usuario do PO ao lado do dono quando souber o usuario dele
*       @antoniobt12062002
/docs/  @antoniobt12062002
```

`.github/pull_request_template.md`:
```markdown
## Resumo
<!-- O que muda e por que -->

## Tipo de mudanca
- [ ] feat
- [ ] fix
- [ ] docs
- [ ] chore / ci / refactor

## Checklist
- [ ] Titulo no padrao Conventional Commits (`feat: ...`, `fix: ...`, `docs: ...`)
- [ ] Docs atualizados
- [ ] Testes adicionados/atualizados (quando houver codigo)

## O que o PO precisa validar
<!-- Pontos de negocio para revisao -->
```

`.github/dependabot.yml`:
```yaml
version: 2
updates:
  - package-ecosystem: github-actions
    directory: /
    schedule:
      interval: weekly
    commit-message:
      prefix: "ci"
```

- [ ] **Step 2: Verify markdownlint passes on existing docs**

Run: `npx --yes markdownlint-cli2 "**/*.md" "#node_modules" "#docs/superpowers"`
Expected: `Summary: 0 error(s)`. If rules fire, add them as `false` in `.markdownlint.json` and re-run.

- [ ] **Step 3: Commit and push scripts + files**

```bash
git add VERSION CHANGELOG.md .markdownlint.json .github
git commit -m "chore: adiciona VERSION, CHANGELOG, lint e templates do GitHub

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>"
git push origin main
```

## Chunk 2: Workflows

### Task 5: `ci.yml`

**Files:**
- Create: `.github/workflows/ci.yml`

- [ ] **Step 1: Write the workflow**

```yaml
name: CI

on:
  pull_request:
    types: [opened, edited, synchronize, reopened]
    branches: [develop, main, 'release/**']
  workflow_dispatch:

concurrency:
  group: ci-${{ github.event.pull_request.number || github.ref }}
  cancel-in-progress: true

permissions:
  contents: read

jobs:
  pr-title:
    if: github.event_name == 'pull_request'
    runs-on: ubuntu-latest
    permissions:
      pull-requests: read
    steps:
      - uses: amannn/action-semantic-pull-request@v5
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  changes:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: read
    outputs:
      docs: ${{ steps.out.outputs.docs }}
      api: ${{ steps.out.outputs.api }}
      web: ${{ steps.out.outputs.web }}
    steps:
      - uses: actions/checkout@v4
      - uses: dorny/paths-filter@v3
        id: f
        if: github.event_name == 'pull_request'
        with:
          filters: |
            docs:
              - 'docs/**'
              - '**/*.md'
              - '.markdownlint.json'
            api:
              - 'api/**'
            web:
              - 'web/**'
      - id: out
        env:
          EVENT: ${{ github.event_name }}
          DOCS: ${{ steps.f.outputs.docs }}
          API: ${{ steps.f.outputs.api }}
          WEB: ${{ steps.f.outputs.web }}
        run: |
          if [ "$EVENT" = "pull_request" ]; then
            echo "docs=$DOCS" >> "$GITHUB_OUTPUT"
            echo "api=$API" >> "$GITHUB_OUTPUT"
            echo "web=$WEB" >> "$GITHUB_OUTPUT"
          else
            echo "docs=true" >> "$GITHUB_OUTPUT"
            echo "api=true" >> "$GITHUB_OUTPUT"
            echo "web=true" >> "$GITHUB_OUTPUT"
          fi

  docs:
    needs: changes
    if: needs.changes.outputs.docs == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: DavidAnson/markdownlint-cli2-action@v16
        with:
          globs: '**/*.md'
      - uses: lycheeverse/lychee-action@v2
        with:
          args: --offline --no-progress './**/*.md'
          fail: true

  api:
    needs: changes
    if: needs.changes.outputs.api == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Placeholder
        run: echo "TODO (quando existir /api): go vet ./... && go test ./... && go build ./..."

  web:
    needs: changes
    if: needs.changes.outputs.web == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Placeholder
        run: echo "TODO (quando existir /web): npm ci && npm run lint && npx tsc --noEmit && npm run build"

  ci-gate:
    if: always()
    needs: [pr-title, changes, docs, api, web]
    runs-on: ubuntu-latest
    steps:
      - name: Falha se algum job falhou ou foi cancelado
        env:
          RESULTS: ${{ join(needs.*.result, ' ') }}
        run: |
          echo "Resultados: $RESULTS"
          case " $RESULTS " in
            *" failure "*|*" cancelled "*) exit 1 ;;
          esac
```

- [ ] **Step 2: Validate YAML syntax**

Run: `python -c "import yaml,sys; yaml.safe_load(open('.github/workflows/ci.yml')); print('ok')"`
Expected: `ok` (if `yaml` is missing, use `npx --yes actionlint` or skip; real validation happens in Task 10).

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: adiciona workflow de CI com ci-gate

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>"
```

### Task 6: `deploy-staging.yml`

**Files:**
- Create: `.github/workflows/deploy-staging.yml`

- [ ] **Step 1: Write the workflow**

```yaml
name: Deploy Staging

on:
  push:
    branches: ['release/**']
  workflow_call:
    inputs:
      ref:
        description: Branch/tag a publicar
        type: string
        required: false

concurrency:
  group: deploy-staging
  cancel-in-progress: false

permissions:
  contents: read

jobs:
  deploy:
    runs-on: ubuntu-latest
    environment: staging
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ inputs.ref || github.ref }}
      - name: Versao
        run: echo "Publicando $(cat VERSION) em staging"
      - name: Placeholder de deploy
        run: echo "TODO: build da API (imagem Docker) + front e deploy no provedor escolhido; segredos via GitHub Secrets do ambiente staging"
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/deploy-staging.yml
git commit -m "ci: adiciona workflow de deploy de staging

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>"
```

### Task 7: `cut-release.yml`

**Files:**
- Create: `.github/workflows/cut-release.yml`

- [ ] **Step 1: Write the workflow**

```yaml
name: Cut release

on:
  workflow_dispatch:
    inputs:
      bump:
        description: Tipo de bump (auto calcula pelos commits)
        type: choice
        options: [auto, major, minor, patch]
        default: auto

permissions:
  contents: write

jobs:
  cut:
    if: github.ref == 'refs/heads/develop'
    runs-on: ubuntu-latest
    outputs:
      version: ${{ steps.v.outputs.version }}
      branch: ${{ steps.v.outputs.branch }}
    steps:
      - uses: actions/checkout@v4
        with:
          ref: develop
          fetch-depth: 0
      - name: Calcula versao
        id: v
        run: |
          version=$(bash .github/scripts/next-version.sh "${{ inputs.bump }}")
          echo "version=$version" >> "$GITHUB_OUTPUT"
          echo "branch=release/v$version" >> "$GITHUB_OUTPUT"
      - name: Cria release branch, VERSION e CHANGELOG
        run: |
          version="${{ steps.v.outputs.version }}"
          git config user.name "github-actions[bot]"
          git config user.email "41898282+github-actions[bot]@users.noreply.github.com"
          git checkout -b "release/v$version"
          bash .github/scripts/changelog-entry.sh "$version" > /tmp/entry.md
          { head -n 2 CHANGELOG.md; cat /tmp/entry.md; tail -n +3 CHANGELOG.md; } > /tmp/CHANGELOG.md
          mv /tmp/CHANGELOG.md CHANGELOG.md
          echo "$version" > VERSION
          git add VERSION CHANGELOG.md
          git commit -m "chore(release): v$version"
          git push origin "release/v$version"

  deploy-staging:
    needs: cut
    uses: ./.github/workflows/deploy-staging.yml
    with:
      ref: ${{ needs.cut.outputs.branch }}
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/cut-release.yml
git commit -m "ci: adiciona workflow para cortar release

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>"
```

### Task 8: `release.yml`

**Files:**
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: Write the workflow**

```yaml
name: Release

on:
  push:
    branches: [main]

permissions:
  contents: write
  pull-requests: write
  actions: write

jobs:
  release:
    runs-on: ubuntu-latest
    outputs:
      version: ${{ steps.v.outputs.version }}
      created: ${{ steps.tag.outputs.created }}
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - id: v
        run: echo "version=$(tr -d '[:space:]' < VERSION)" >> "$GITHUB_OUTPUT"
      - id: tag
        env:
          GH_TOKEN: ${{ github.token }}
        run: |
          version="${{ steps.v.outputs.version }}"
          if git rev-parse "v$version" >/dev/null 2>&1; then
            echo "Tag v$version ja existe; nada a fazer."
            echo "created=false" >> "$GITHUB_OUTPUT"
            exit 0
          fi
          awk -v h="## v$version " 'index($0,h)==1{f=1;next} /^## v/{f=0} f' CHANGELOG.md > notes.md
          gh release create "v$version" --target "$GITHUB_SHA" --title "v$version" --notes-file notes.md
          echo "created=true" >> "$GITHUB_OUTPUT"

  deploy-production:
    needs: release
    if: needs.release.outputs.created == 'true'
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4
      - name: Placeholder de deploy
        run: echo "TODO: deploy de $(cat VERSION) em producao; aprovacao do PO via Environment production"

  back-merge:
    needs: release
    if: needs.release.outputs.created == 'true'
    runs-on: ubuntu-latest
    env:
      GH_TOKEN: ${{ github.token }}
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - name: Abre PR main -> develop e dispara o CI na branch
        run: |
          version="${{ needs.release.outputs.version }}"
          branch="backmerge/v$version"
          git checkout -b "$branch"
          git push origin "$branch"
          gh pr create --base develop --head "$branch" \
            --title "chore: back-merge v$version na develop" \
            --body "Traz a release v$version de volta para a develop. **Use 'Create a merge commit' (nao squash)** para a tag continuar alcancavel pela develop."
          gh workflow run ci.yml --ref "$branch"
```

- [ ] **Step 2: Commit and push all workflows**

```bash
git add .github/workflows/release.yml
git commit -m "ci: adiciona workflow de release, deploy de producao e back-merge

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>"
git push origin main
```

### Task 9: `CONTRIBUTING.md`

**Files:**
- Create: `docs/CONTRIBUTING.md`

- [ ] **Step 1: Write the doc**

```markdown
# Como contribuir

## Branches

- `main`: producao. `develop`: integracao (branch padrao). `release/vX.Y.Z`: versao em homologacao (staging).
- Trabalhe em `feature/<assunto>` a partir da `develop` e abra PR para `develop` (squash).
- Correcao urgente em producao: `hotfix/<assunto>` a partir da `main`, PR para `main`, depois back-merge na `develop`.

## Commits e titulo de PR

Padrao Conventional Commits: `feat: ...`, `fix: ...`, `docs: ...`, `chore: ...`. Use `feat!:` ou `BREAKING CHANGE:` para mudanca incompativel. O titulo do PR vira a mensagem do squash e define a versao.

## Cortando uma release

1. Actions > **Cut release** > Run workflow na branch `develop` (bump `auto` ou forcado).
2. O workflow cria `release/vX.Y.Z`, atualiza `VERSION`/`CHANGELOG.md` e publica em **staging**.
3. Usuario final valida em staging; correcoes entram na release via PR.
4. Abra o PR: `gh pr create --base main --head release/vX.Y.Z --title "chore(release): vX.Y.Z"` e faca merge com **merge commit**.
5. O workflow **Release** cria a tag e a GitHub Release, pede aprovacao do PO para producao e abre o PR de back-merge `main` -> `develop`.
6. Faca o merge do back-merge com **Create a merge commit** (nunca squash) **antes** da proxima release.

## Segredos

O repositorio e publico: nunca comite segredos. Use GitHub Secrets por ambiente (`staging`, `production`).
```

- [ ] **Step 2: Lint and commit**

```bash
npx --yes markdownlint-cli2 "docs/CONTRIBUTING.md"
git add docs/CONTRIBUTING.md
git commit -m "docs: adiciona guia de contribuicao e fluxo de release

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>"
git push origin main
```

## Chunk 3: Repository settings and end-to-end validation

### Task 10: Create `develop`, settings, rulesets, environments

**Files:**
- Create: `.github/rulesets/protect-main-develop.json`, `.github/rulesets/protect-release.json`

- [ ] **Step 1: Create `develop` from `main` and set it as default**

```bash
git push origin main:develop
gh repo edit antoniobt12062002/torcida-jovem-platform --default-branch develop
```

- [ ] **Step 2: Merge settings and Actions PR permissions**

```bash
R=antoniobt12062002/torcida-jovem-platform
gh api -X PATCH repos/$R -F allow_squash_merge=true -F allow_merge_commit=true -F allow_rebase_merge=false \
  -F delete_branch_on_merge=true -f squash_merge_commit_title=PR_TITLE -f squash_merge_commit_message=BLANK
gh api -X PUT repos/$R/actions/permissions/workflow -f default_workflow_permissions=read -F can_approve_pull_request_reviews=true
```

- [ ] **Step 3: Write the ruleset files**

`.github/rulesets/protect-main-develop.json`:
```json
{
  "name": "protect-main-develop",
  "target": "branch",
  "enforcement": "active",
  "conditions": { "ref_name": { "include": ["refs/heads/main", "refs/heads/develop"], "exclude": [] } },
  "bypass_actors": [{ "actor_id": 5, "actor_type": "RepositoryRole", "bypass_mode": "pull_request" }],
  "rules": [
    { "type": "deletion" },
    { "type": "non_fast_forward" },
    { "type": "pull_request", "parameters": {
        "required_approving_review_count": 1,
        "dismiss_stale_reviews_on_push": true,
        "require_code_owner_review": false,
        "require_last_push_approval": false,
        "required_review_thread_resolution": false } },
    { "type": "required_status_checks", "parameters": {
        "strict_required_status_checks_policy": true,
        "required_status_checks": [{ "context": "ci-gate" }] } }
  ]
}
```

`.github/rulesets/protect-release.json`:
```json
{
  "name": "protect-release",
  "target": "branch",
  "enforcement": "active",
  "conditions": { "ref_name": { "include": ["refs/heads/release/**"], "exclude": [] } },
  "rules": [{ "type": "deletion" }, { "type": "non_fast_forward" }]
}
```

- [ ] **Step 4: Apply the rulesets**

```bash
R=antoniobt12062002/torcida-jovem-platform
gh api -X POST repos/$R/rulesets --input .github/rulesets/protect-main-develop.json
gh api -X POST repos/$R/rulesets --input .github/rulesets/protect-release.json
gh api repos/$R/rulesets --jq '.[] | {name, enforcement}'
```
Expected: two rulesets, both `active`.

- [ ] **Step 5: Create Environments** (owner as production reviewer until the PO's user id is known)

```bash
R=antoniobt12062002/torcida-jovem-platform
gh api -X PUT repos/$R/environments/staging
echo '{"reviewers":[{"type":"User","id":82279367}]}' | gh api -X PUT repos/$R/environments/production --input -
gh api repos/$R/environments --jq '.environments[].name'
```
Expected: `staging` and `production`.

- [ ] **Step 6: Commit ruleset definitions via PR** (branches are protected now)

```bash
git checkout -b chore/rulesets develop
git add .github/rulesets
git commit -m "chore: versiona definicao dos rulesets

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>"
git push -u origin chore/rulesets
gh pr create --base develop --title "chore: versiona definicao dos rulesets" --body "Guarda o JSON dos rulesets aplicados via gh api."
```
Expected: `ci-gate` runs and passes; merge as admin (bypass) with squash.

### Task 11: Retarget PR #1 and end-to-end test

- [ ] **Step 1: Fix PR #1 title and base**

```bash
gh pr edit 1 --base develop --title "docs: adiciona documentacao inicial do produto"
gh pr update-branch 1 || true
gh pr checks 1 --watch
```
Expected: `pr-title`, `docs`, `ci-gate` pass. If `docs` fails on lint/links, fix on the PR branch.

- [ ] **Step 2: Merge PR #1** (squash; admin bypass for the 1 approval unless the PO already approved).

- [ ] **Step 3: Confirm with the user before creating a real tag**, then cut a release

```bash
gh workflow run cut-release.yml --ref develop -f bump=minor
gh run watch
git fetch origin && git branch -r | grep release/
```
Expected: `origin/release/v0.1.0`, `deploy-staging` job green (placeholder), `VERSION` = `0.1.0`, `CHANGELOG.md` has a `## v0.1.0` section.

- [ ] **Step 3b: Open and merge the release PR**

```bash
gh pr create --base main --head release/v0.1.0 --title "chore(release): v0.1.0" --body "Release v0.1.0"
gh pr checks --watch
gh pr merge --merge --admin
```
Expected: `release.yml` creates tag `v0.1.0` + GitHub Release, `deploy-production` waits for approval (approve in the Actions UI), `back-merge` opens `backmerge/v0.1.0` -> `develop` with `ci-gate` reported by the dispatched run.

- [ ] **Step 4: Merge the back-merge PR with a merge commit**

```bash
gh pr merge backmerge/v0.1.0 --merge --admin
git fetch origin --tags && git checkout develop && git pull
git describe --tags --abbrev=0
```
Expected: `v0.1.0` reachable from `develop`.

- [ ] **Step 5: Record findings** — if `cut-release` could not push the branch, or the back-merge PR lacks `ci-gate`, note the fix needed (e.g. bypass actor for github-actions, or PAT) and address it before declaring done.

### Task 12: Align spec and open PR for spec + plan

- [ ] **Step 1: Update the spec** on `docs/spec-github-cicd` (rebased on `develop`) so sections 3, 5 and 6 reflect the deviations listed at the top of this plan (no `pr-title.yml`, `release/*` rules, merge-commit back-merge, `.github/rulesets/`).

- [ ] **Step 2: Open the PR**

```bash
git checkout docs/spec-github-cicd && git rebase develop
git add docs/superpowers
git commit -m "docs: alinha spec e adiciona plano de implementacao do CI/CD

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>"
git push -u origin docs/spec-github-cicd
gh pr create --base develop --title "docs: spec e plano do fluxo de branches e pipelines" --body "Spec e plano de implementacao do CI/CD."
```
Expected: `ci-gate` green.
