# Web: como rodar o front com a API

Guia para subir a Web (`web/`, Next.js 16) junto com a API Go, o PostgreSQL e o armazenamento S3 local, e para rodar os gates da Web. Decisões de fundo: [ADR-010](../adr/010-acesso-do-navegador-a-api-por-rewrite-same-origin.md) (rewrite same-origin) e `.specs/features/web/STATE.md`. A verificação original contra a API real está em `.specs/features/web/evidence/f1-rewrite.md`.

Os valores deste guia são só de desenvolvimento local. Não há senha, chave nem cookie reais aqui: onde aparece `<...>`, use um valor seu, gerado na hora e nunca commitado.

## Pré-requisitos

- Docker (PostgreSQL 16 pelo `docker-compose.yml` e o contêiner S3 avulso).
- Go (versão do `api/go.mod`) para a API e o `bootstrap-admin`.
- Node 22.12 ou mais novo e pnpm (versão do campo `packageManager` de `web/package.json`).

## Variáveis

### API

A API lê as variáveis do ambiente do processo; ela **não** carrega o `.env` sozinha. O `.env.example` da raiz documenta todas. As que importam para a Web:

| Variável | Valor local | Por quê |
| --- | --- | --- |
| `APP_ENV` | `development` | Libera `COOKIE_SECURE=false` e os padrões de desenvolvimento. |
| `DATABASE_URL` | `postgres://tj_app:tj_app_dev@localhost:5432/tj?sslmode=disable` | Papel de aplicação do banco do compose. |
| `ALLOWED_ORIGINS` | `http://localhost:3000` | A origem do front. A API recusa escrita de outra origem (`403 origin_not_allowed`). Com o rewrite, o navegador só fala com essa origem. |
| `COOKIE_SECURE` | `false` | Sem HTTPS local o navegador não guardaria um cookie `Secure`. Só é aceito com `APP_ENV=development`. |
| `COOKIE_DOMAIN` | vazio | O cookie `tj_session` fica host-only na origem do front. |
| `APP_BASE_URL` | `http://localhost:3000` | Base dos links enviados por e-mail (por exemplo `/redefinir-senha#token=`): é o endereço do **front**, não o da API. Fora de development, exige `https`. |
| `EMAIL_PROVIDER` | `log` | Registra só o domínio do destinatário e o assunto. O link de recuperação **não** aparece no log, então a redefinição com token válido não é exercitável localmente. |
| `STORAGE_ENABLED` | `true` | **Obrigatório**: a API não sobe sem armazenamento (`STORAGE_ENABLED=true é obrigatório`), porque o financeiro anexa comprovantes. O `.env.example` ainda traz `false`; troque no seu `.env`. |
| `S3_ENDPOINT` | `http://localhost:3900` | Endereço do S3 local (seção "Armazenamento S3 local"). |
| `S3_REGION` | `garage` | Região configurada no `garage.toml`. |
| `S3_BUCKET` | `tj-local` | Bucket criado pelo Garage na subida. |
| `S3_ACCESS_KEY`, `S3_SECRET_KEY` | `<chave local>`, `<segredo local>` | Os mesmos valores passados ao contêiner Garage. Gere na hora; nunca commite. |
| `S3_USE_PATH_STYLE` | `true` | Exigido pelo emulador local. |

### Front

| Variável | Valor local | Comportamento |
| --- | --- | --- |
| `API_URL` | `http://localhost:8080` | Destino do rewrite de `/api/v1/*`. **Sem valor padrão**: `next dev`, `next build` e `next typegen` falham sem ela (`API_URL não está definida: ...`). **Fixado no build**: `next build` grava o destino em `.next/routes-manifest.json` e `next start` usa esse valor, ignorando o `API_URL` do ambiente em execução. Mudar o endereço da API exige novo build. |

No CI, o job `web` fornece `API_URL=http://localhost:8080` só nos passos `next typegen` e `next build`, como valor de validação (não há API rodando no CI). Cada ambiente implantado fornece o seu próprio `API_URL` no build.

## Ordem de subida

Todos os comandos partem da raiz do repositório, em bash.

### 1. Banco e migrações

```bash
cp .env.example .env    # ignorado pelo git; ajuste STORAGE_ENABLED e S3_* como acima
docker compose up -d db
docker compose --profile tools run --rm migrate
```

Para recriar o banco do zero: `docker compose down -v`.

### 2. Armazenamento S3 local (Garage avulso)

O `docker-compose.yml` não tem serviço S3. Use um contêiner Garage avulso, com a mesma imagem e configuração dos testes de integração (`api/internal/platform/testutil/s3.go`). A configuração fica num diretório temporário fora do repositório.

```bash
GARAGE_DIR="$(mktemp -d)"
cat > "$GARAGE_DIR/garage.toml" <<EOF
metadata_dir = "/tmp/garage/meta"
data_dir = "/tmp/garage/data"
db_engine = "sqlite"
replication_factor = 1

rpc_bind_addr = "[::]:3901"
rpc_public_addr = "127.0.0.1:3901"
rpc_secret = "$(openssl rand -hex 32)"

[s3_api]
s3_region = "garage"
api_bind_addr = "[::]:3900"
root_domain = ".s3.garage.localhost"
EOF

export S3_ACCESS_KEY="GK$(openssl rand -hex 12)"
export S3_SECRET_KEY="$(openssl rand -hex 32)"

docker run -d --name tj-garage -p 3900:3900 \
  -v "$GARAGE_DIR/garage.toml:/etc/garage.toml:ro" \
  -e GARAGE_DEFAULT_ACCESS_KEY="$S3_ACCESS_KEY" \
  -e GARAGE_DEFAULT_SECRET_KEY="$S3_SECRET_KEY" \
  -e GARAGE_DEFAULT_BUCKET=tj-local \
  dxflrs/garage:v2.4.1 /garage server --single-node --default-bucket
```

Os dados ficam dentro do contêiner: `docker rm -f tj-garage` apaga tudo. A porta abre um pouco antes de o bucket ficar pronto; se a primeira escrita de comprovante falhar logo após a subida, espere alguns segundos.

No Git Bash do Windows, se o Docker não achar o arquivo montado, use o caminho no formato do Windows (`cygpath -w "$GARAGE_DIR"`) e prefixe o comando com `MSYS_NO_PATHCONV=1`.

### 3. Primeiro administrador

Com o banco migrado, crie a primeira conta pelo `bootstrap-admin` (detalhes em `docs/CONTRIBUTING.md`, "Primeiro administrador"). A senha vai **só** pela variável `BOOTSTRAP_ADMIN_PASSWORD`, nunca por opção de linha de comando:

```bash
cd api
export DATABASE_URL="postgres://tj_app:tj_app_dev@localhost:5432/tj?sslmode=disable"
read -rs BOOTSTRAP_ADMIN_PASSWORD && export BOOTSTRAP_ADMIN_PASSWORD
go run ./cmd/bootstrap-admin --email presidente@local.test --name "Presidente Local" --role PRESIDENTE
unset BOOTSTRAP_ADMIN_PASSWORD
cd ..
```

O padrão é `ADMIN_SISTEMA`, que administra usuários mas não opera `financeiro` nem `estoque`. A API só deixa alguém conceder um papel cujas permissões ele mesmo tem: um `ADMIN_SISTEMA` que tenta promover alguém a `TESOURARIA`, `ESTOQUE_LOJA` ou `CONSELHO_FISCAL` recebe `403 privilege_escalation`. Como o `bootstrap-admin` só cria a primeira conta administrativa, para exercitar localmente o ciclo completo crie-a com `--role PRESIDENTE`, que recebe todas as permissões do catálogo. Depois crie as pessoas pela tela de usuários (`/admin/usuarios`) e promova-as com o papel certo (o motivo da promoção precisa de pelo menos 10 caracteres). O primeiro login de qualquer conta exige a troca da senha.

### 4. API

```bash
cd api
export APP_ENV=development COOKIE_SECURE=false \
  DATABASE_URL="postgres://tj_app:tj_app_dev@localhost:5432/tj?sslmode=disable" \
  ALLOWED_ORIGINS=http://localhost:3000 APP_BASE_URL=http://localhost:3000 EMAIL_PROVIDER=log \
  STORAGE_ENABLED=true S3_ENDPOINT=http://localhost:3900 S3_REGION=garage S3_BUCKET=tj-local \
  S3_USE_PATH_STYLE=true
# S3_ACCESS_KEY e S3_SECRET_KEY já exportadas no passo 2
go run ./cmd/api
```

Confira com `curl http://localhost:8080/healthz` (`{"status":"ok"}`).

Não carregue o `.env` com `source .env`: a linha `EMAIL_FROM=TJ Platform <no-reply@localhost>` não é válida em shell. Exporte as variáveis como acima.

### 5. Front

Em outro terminal:

```bash
cd web
pnpm install
API_URL=http://localhost:8080 pnpm dev
```

Abra `http://localhost:3000`: a raiz leva a `/inicio` e, sem sessão, a `/entrar`.

Para rodar como em produção (é como o smoke manual é feito):

```bash
API_URL=http://localhost:8080 pnpm build
pnpm start    # usa o API_URL gravado no build
```

### Desmontagem

Pare a API e o Next (`Ctrl+C`), depois:

```bash
docker rm -f tj-garage
docker compose down -v
rm -rf "$GARAGE_DIR"
```

## Rewrite same-origin

O navegador só fala com a origem do front (`http://localhost:3000`). O `web/next.config.ts` encaminha `/api/v1/*` para `${API_URL}/api/v1/*` (WEB-D-001, ADR-010). Consequências:

- O cookie `tj_session` (`HttpOnly`, `SameSite=Lax`) chega como cookie da origem do front e nunca é lido pelo JavaScript.
- As escritas saem com o `Origin` do front e com o cabeçalho `X-CSRF-Token`, cujo valor vem de `GET /api/v1/auth/me`. Sem ele a API responde `403 csrf_invalid`; com outra origem, `403 origin_not_allowed`.
- Sem CORS e sem mudança na API. Não há `proxy.ts`: a proteção de rotas fica no layout autenticado (`web/app/(app)/layout.tsx`), e a API continua sendo a autoridade de acesso.
- Chame sempre a API pelo `:3000` quando estiver testando o front. Chamar o `:8080` direto pula o rewrite e não reproduz o que o navegador faz.

### Limite de upload

O rewrite do Next trunca corpos acima de cerca de 10 MiB: o upload chega truncado à API e a pessoa recebe `500` em texto, não o `413 document_too_large` da API. Por isso o front recusa, **antes do envio**, comprovantes acima de **10 MiB − 64 KiB (10.420.224 bytes)**, deixando margem para o envelope `multipart/form-data` (`web/features/financeiro/comprovantes/limits.ts`). O limite da API continua 10 MiB e é a autoridade. Não se usa `experimental.proxyClientMaxBodySize` nem proxy externo na V1 (ADR-010, Notas; WEB-D-009).

## Gates da Web

A partir de `web/`, com `API_URL` no ambiente (exigido por `next typegen` e `next build`):

```bash
export API_URL=http://localhost:8080
pnpm install --frozen-lockfile
pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build   # Web
pnpm audit --audit-level high                                                              # Audit
pnpm lint:api && pnpm gen:api:check                                                        # Contrato
```

- `next typegen` vem antes do `tsc`: sem ele, os tipos de rota (`LayoutProps`, `PageProps`) não existem e o `tsc` falha.
- `pnpm test` roda o Vitest com MSW (`onUnhandledRequest: "error"`): toda chamada à API nos testes precisa de um handler.
- O contrato OpenAPI é a fonte dos tipos (`web/lib/api/*.d.ts`); a Web não altera `api/openapi/*.yaml`. Se o contrato mudar, `pnpm gen:api` regenera e `pnpm gen:api:check` reprova diferença.
- No Windows (PowerShell), para capturar o código de saída real de cada passo: `cmd /v:on /c "pnpm test & echo EXIT=!ERRORLEVEL!"`.

## Problemas comuns

| Sintoma | Causa |
| --- | --- |
| `next build`/`next dev` falha com `API_URL não está definida` | Falta `API_URL` no ambiente. |
| O front chama a API errada depois de mudar `API_URL` | O destino foi fixado no build anterior; rode `pnpm build` de novo. |
| A API não sobe: `STORAGE_ENABLED=true é obrigatório` | `.env` copiado do exemplo; ligue o armazenamento e suba o Garage. |
| Login funciona, mas toda escrita dá `403 origin_not_allowed` | `ALLOWED_ORIGINS` não contém a origem do front. |
| Login responde `200` mas a sessão não fica | `COOKIE_SECURE=true` sem HTTPS local. |
| Promover alguém a `TESOURARIA` dá `403 privilege_escalation` | A conta que promove não tem as permissões do papel (por exemplo, `ADMIN_SISTEMA`); use um `PRESIDENTE`. |
| Upload de comprovante dá `500` | Anexo grande demais passou pelo rewrite sem a checagem do front. |
