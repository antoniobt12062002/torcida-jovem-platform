# Scaffold de /api (Go) e /web (Next.js) Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Esqueleto do monorepo rodando de ponta a ponta (Postgres local, API Gin com `/healthz`, front Next.js mostrando o status) e CI real nos jobs `api` e `web`.

**Architecture:** API em Go (config, database, httpapi) com ping injetável para teste; migrações SQL com golang-migrate via docker compose; front Next.js App Router que consulta a API no servidor (`API_URL`, sem CORS).

**Tech Stack:** Go, Gin, GORM (driver postgres), Postgres 16, golang-migrate, Next.js + TypeScript + Tailwind + shadcn/ui, pnpm, GitHub Actions.

**Spec:** `docs/superpowers/specs/2026-09-26-scaffold-api-web-design.md`

**Branch:** `feature/scaffold-api-web` (a partir de `develop`); tudo entra por PR para `develop`.

---

## Chunk 1: API

### Task 1: Módulo Go e config (TDD)

**Files:** Create `api/go.mod`, `api/internal/config/config.go`, `api/internal/config/config_test.go`

- [ ] **Step 1: Iniciar módulo**

```bash
mkdir -p api && cd api && go mod init github.com/antoniobt12062002/torcida-jovem-platform/api
```

- [ ] **Step 2: Teste que falha** — `api/internal/config/config_test.go`

```go
package config

import "testing"

func TestLoad(t *testing.T) {
	t.Run("usa porta padrao", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://x")
		t.Setenv("PORT", "")
		cfg, err := Load()
		if err != nil || cfg.Port != "8080" || cfg.DatabaseURL != "postgres://x" {
			t.Fatalf("cfg=%+v err=%v", cfg, err)
		}
	})
	t.Run("respeita PORT", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://x")
		t.Setenv("PORT", "9000")
		cfg, _ := Load()
		if cfg.Port != "9000" {
			t.Fatalf("port=%s", cfg.Port)
		}
	})
	t.Run("exige DATABASE_URL", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "")
		if _, err := Load(); err == nil {
			t.Fatal("esperava erro")
		}
	})
}
```

- [ ] **Step 3: Rodar e ver falhar** — `cd api && go test ./internal/config/` → FAIL (`undefined: Load`).

- [ ] **Step 4: Implementar** — `api/internal/config/config.go`

```go
package config

import (
	"errors"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
}

func Load() (Config, error) {
	cfg := Config{Port: os.Getenv("PORT"), DatabaseURL: os.Getenv("DATABASE_URL")}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL e obrigatoria")
	}
	return cfg, nil
}
```

- [ ] **Step 5: Rodar** — `go test ./internal/config/` → PASS.
- [ ] **Step 6: Commit** — `git add api && git commit -m "feat(api): adiciona modulo Go e carregamento de config"`

### Task 2: Router e /healthz (TDD)

**Files:** Create `api/internal/httpapi/router.go`, `api/internal/httpapi/router_test.go`

- [ ] **Step 1: Teste que falha** — `router_test.go`

```go
package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	cases := []struct {
		name string
		ping PingFunc
		code int
		body string
	}{
		{"banco ok", func(context.Context) error { return nil }, http.StatusOK, `{"status":"ok"}`},
		{"banco fora", func(context.Context) error { return errors.New("down") }, http.StatusServiceUnavailable, `{"status":"degraded"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			NewRouter(tc.ping).ServeHTTP(w, req)
			if w.Code != tc.code || w.Body.String() != tc.body {
				t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
}
```

- [ ] **Step 2:** `go get github.com/gin-gonic/gin && go test ./internal/httpapi/` → FAIL (`undefined: PingFunc`).

- [ ] **Step 3: Implementar** — `router.go`

```go
package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type PingFunc func(ctx context.Context) error

func NewRouter(ping PingFunc) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
	r.GET("/healthz", healthz(ping))
	return r
}

func healthz(ping PingFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
```

- [ ] **Step 4:** `go test ./internal/httpapi/` → PASS.
- [ ] **Step 5: Commit** — `git add api && git commit -m "feat(api): adiciona router Gin e endpoint /healthz"`

### Task 3: Banco, main, migração

**Files:** Create `api/internal/database/database.go`, `api/cmd/api/main.go`, `api/migrations/000001_init.up.sql`, `api/migrations/000001_init.down.sql`

- [ ] **Step 1: `database.go`**

```go
package database

import (
	"context"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Open(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func Ping(db *gorm.DB) func(context.Context) error {
	return func(ctx context.Context) error {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.PingContext(ctx)
	}
}
```

- [ ] **Step 2: `cmd/api/main.go`**

```go
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/config"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/database"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/httpapi"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.NewRouter(database.Ping(db)),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **Step 3: Migração** — `000001_init.up.sql`: `CREATE EXTENSION IF NOT EXISTS pgcrypto;` — `000001_init.down.sql`: `DROP EXTENSION IF EXISTS pgcrypto;` (habilita `gen_random_uuid()` para os ids das entidades futuras).

- [ ] **Step 4:** `cd api && go mod tidy && go vet ./... && go test ./... && go build ./...` → tudo ok.
- [ ] **Step 5: Commit** — `git add api && git commit -m "feat(api): adiciona conexao GORM, main com graceful shutdown e migracao inicial"`

### Task 4: Dockerfile, lint, compose, env, verificação local

**Files:** Create `api/Dockerfile`, `api/.golangci.yml`, `docker-compose.yml`, `.env.example`; Modify `.gitignore`

- [ ] **Step 1: `api/Dockerfile`** (usar a versão de `go.mod` na tag, ex.: `golang:1.26`)

```dockerfile
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
EXPOSE 8080
ENTRYPOINT ["/api"]
```

- [ ] **Step 2: `api/.golangci.yml`**

```yaml
version: "2"
linters:
  enable:
    - misspell
```

- [ ] **Step 3: `docker-compose.yml`** (credenciais apenas de desenvolvimento local)

```yaml
services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: tj
      POSTGRES_PASSWORD: tj
      POSTGRES_DB: tj
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U tj -d tj"]
      interval: 3s
      timeout: 3s
      retries: 10

  migrate:
    image: migrate/migrate:v4.18.1
    profiles: [tools]
    depends_on:
      db:
        condition: service_healthy
    volumes:
      - ./api/migrations:/migrations
    command: ["-path=/migrations", "-database=postgres://tj:tj@db:5432/tj?sslmode=disable", "up"]

volumes:
  pgdata:
```

- [ ] **Step 4: `.env.example`**

```
DATABASE_URL=postgres://tj:tj@localhost:5432/tj?sslmode=disable
PORT=8080
API_URL=http://localhost:8080
```

- [ ] **Step 5: `.gitignore`** — acrescentar `.env`, `.env.local`, `node_modules/`, `.next/`.

- [ ] **Step 6: Verificar de ponta a ponta**

```bash
docker compose up -d db && docker compose --profile tools run --rm migrate
cd api && DATABASE_URL="postgres://tj:tj@localhost:5432/tj?sslmode=disable" go run ./cmd/api &
curl -s localhost:8080/healthz          # {"status":"ok"}
docker compose stop db; curl -s -o /dev/null -w "%{http_code}" localhost:8080/healthz   # 503
docker build -t tj-api ./api            # builda a imagem
```
Expected: `{"status":"ok"}`, depois `503`, e build da imagem concluído. Encerrar a API e `docker compose down`.

- [ ] **Step 7: Commit** — `git add api docker-compose.yml .env.example .gitignore && git commit -m "feat(api): adiciona Dockerfile, lint, docker compose e env de exemplo"`

## Chunk 2: Web

### Task 5: Next.js + shadcn + página de status

**Files:** Create `web/**`

- [ ] **Step 1: Gerar o app**

```bash
pnpm create next-app@latest web --ts --tailwind --eslint --app --no-src-dir --use-pnpm --import-alias "@/*" --yes
cd web && echo "22" > .nvmrc && pnpm pkg set packageManager=pnpm@10.30.3
pnpm dlx shadcn@latest init --defaults && pnpm dlx shadcn@latest add badge --yes
```
Se o gerador criar `web/README.md`/`AGENTS.md`, manter (o lint de markdown do CI os ignora, ver Task 6).

- [ ] **Step 2: `web/app/page.tsx`**

```tsx
import { Badge } from "@/components/ui/badge";

export const dynamic = "force-dynamic";

async function apiOnline(): Promise<boolean> {
  try {
    const res = await fetch(`${process.env.API_URL ?? "http://localhost:8080"}/healthz`, {
      cache: "no-store",
    });
    return res.ok;
  } catch {
    return false;
  }
}

export default async function Home() {
  const online = await apiOnline();
  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-4 p-6">
      <h1 className="text-3xl font-bold">TJ Platform</h1>
      <Badge variant={online ? "default" : "destructive"}>
        {online ? "API online" : "API indisponível"}
      </Badge>
    </main>
  );
}
```

- [ ] **Step 3: Verificar** — `pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm build` → ok. Com API e banco de pé, `pnpm start` e abrir `http://localhost:3000` mostra "API online"; com o banco parado, "API indisponível".
- [ ] **Step 4: Commit** — `git add web && git commit -m "feat(web): adiciona app Next.js com pagina de status da API"`

## Chunk 3: CI e documentação

### Task 6: Jobs reais, Dependabot, README

**Files:** Modify `.github/workflows/ci.yml`, `.github/dependabot.yml`, `README.md`

- [ ] **Step 1: `ci.yml` — job `docs`**: trocar os dois passos para ignorar o `web/` (arquivos gerados):

```yaml
      - uses: DavidAnson/markdownlint-cli2-action@v16
        with:
          globs: |
            **/*.md
            !web/**
            !**/node_modules/**
      - uses: lycheeverse/lychee-action@v2
        with:
          args: --offline --no-progress --exclude-path web './**/*.md'
          fail: true
```

- [ ] **Step 2: `ci.yml` — job `api`** (substituir o placeholder)

```yaml
  api:
    needs: changes
    if: needs.changes.outputs.api == 'true'
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: api
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: api/go.mod
          cache-dependency-path: api/go.sum
      - run: go vet ./...
      - uses: golangci/golangci-lint-action@v8
        with:
          version: latest
          working-directory: api
      - run: go test ./...
      - run: go build ./...
```

- [ ] **Step 3: `ci.yml` — job `web`**

```yaml
  web:
    needs: changes
    if: needs.changes.outputs.web == 'true'
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: web
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
        with:
          package_json_file: web/package.json
      - uses: actions/setup-node@v4
        with:
          node-version-file: web/.nvmrc
          cache: pnpm
          cache-dependency-path: web/pnpm-lock.yaml
      - run: pnpm install --frozen-lockfile
      - run: pnpm lint
      - run: pnpm exec next typegen && pnpm exec tsc --noEmit
      - run: pnpm build
```

- [ ] **Step 4: `dependabot.yml`** — acrescentar ecossistemas `gomod` (`/api`) e `npm` (`/web`), ambos `weekly`, com `commit-message.prefix: "chore"` e `open-pull-requests-limit: 3`.

- [ ] **Step 5: `README.md`** — adicionar seção "Desenvolvimento local":

```markdown
## Desenvolvimento local

1. `cp .env.example .env`
2. `docker compose up -d db` e `docker compose --profile tools run --rm migrate`
3. API: `cd api && go run ./cmd/api` (lê `DATABASE_URL` e `PORT` do ambiente)
4. Front: `cd web && pnpm install && pnpm dev`
```

- [ ] **Step 6: Lint e commit** — `npx --yes markdownlint-cli2 README.md`, `npx --yes yaml-lint .github/workflows/ci.yml .github/dependabot.yml`; depois `git add . && git commit -m "ci: ativa jobs reais de api e web, dependabot e README"`.

### Task 7: PR e validação do CI

- [ ] **Step 1:** `git push -u origin feature/scaffold-api-web` e `gh pr create --base develop --title "feat: scaffold de api (Go) e web (Next.js) com CI real"`.
- [ ] **Step 2:** `gh pr checks --watch`. Esperado: `api`, `web`, `docs`, `pr-title` e `ci-gate` verdes. Se `golangci-lint` ou `pnpm build` falharem, corrigir na branch e reenviar.
- [ ] **Step 3:** Merge por squash (com `--admin` enquanto não houver segundo aprovador) e confirmar o CI na `develop`.
