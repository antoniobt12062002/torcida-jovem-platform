# Evidência F1 — rewrite same-origin contra a API real (FND-01 AC5/AC6)

- **Data**: 2026-10-04
- **Unidade**: F1 (tasks `01-fundacao.md`, T3), branch `feature/web-f1`
- **Código verificado**: commit `06d9139` (T2), com o `next.config.ts` de T1 (`rewrites()` de `/api/v1/:path*` para `${API_URL}/api/v1/:path*`)
- **Versões**: Next.js 16.3.8, Node 22.18.0, Go 1.26.1, PostgreSQL 16 (compose), Garage `dxflrs/garage:v2.4.1`

Este arquivo não contém cookie, token CSRF, senha nem chave S3. Os valores de `Set-Cookie` aparecem como `<redigido>`.

## Resultado

| Item de FND-01 AC5 | Resultado |
| --- | --- |
| (a) `tj_session` chega como cookie da origem do front | **Passou**: `200`, `Set-Cookie: tj_session=<redigido>; Path=/; Max-Age=28800; HttpOnly; SameSite=Lax`, sem `Domain` |
| (b) escrita autenticada passa pelo `Origin` | **Passou**: `POST /api/v1/financeiro/contas` → `201 Created` (nunca `403 origin_not_allowed`) |
| (c) comprovante de pelo menos 9,5 MiB chega íntegro | **Passou**: `201 Created`, `size_bytes` = 9961472 = tamanho do arquivo; SHA-256 do arquivo baixado igual ao do enviado |

FND-01 AC6 não foi acionado: os três itens passaram.

Achado fora do AC5, com impacto em decisão aprovada: **o rewrite também limita o corpo a 10 MB** (ver "Limite de corpo de 10 MB no rewrite", abaixo).

## Ambiente montado

Tudo a partir do worktree `tjw/f1`, sem alterar `docker-compose.yml` nem nada em `api/`.

1. `.env` local copiado de `.env.example` (ignorado pelo git: `.gitignore:8:.env`), com `STORAGE_ENABLED=true`, `S3_ENDPOINT=http://localhost:3910`, `S3_REGION=garage`, `S3_BUCKET=<bucket local>`, `S3_USE_PATH_STYLE=true` e as chaves de teste do Garage. Apagado ao fim.
2. Banco: `docker compose up -d db` e `docker compose --profile tools run --rm migrate` (migrações 1 a 9 aplicadas).
3. Armazenamento: contêiner Garage avulso, com a configuração de `api/internal/platform/testutil/s3.go` (mesmo `garage.toml`, `rpc_secret` aleatório, `--single-node --default-bucket`, chave e bucket padrão por `GARAGE_DEFAULT_*`), porta `3910:3900`:

   ```
   docker run -d --name tjw-f1-garage -p 3910:3900 \
     -v <scratch>/garage.toml:/etc/garage.toml:ro \
     -e GARAGE_DEFAULT_ACCESS_KEY=<chave> -e GARAGE_DEFAULT_SECRET_KEY=<segredo> -e GARAGE_DEFAULT_BUCKET=<bucket> \
     dxflrs/garage:v2.4.1 /garage server --single-node --default-bucket
   ```

4. Usuário de teste: `go run ./cmd/bootstrap-admin --email <conta local> --name "Presidente Local" --role PRESIDENTE`, com a senha só em `BOOTSTRAP_ADMIN_PASSWORD` (`docs/CONTRIBUTING.md`, "Primeiro administrador"). `PRESIDENTE` recebe todas as permissões do catálogo, então nenhum papel extra foi atribuído.
5. API: binário de `./cmd/api` com as variáveis do `.env` local (`APP_ENV=development`, `COOKIE_SECURE=false`, `ALLOWED_ORIGINS=http://localhost:3000`, `COOKIE_DOMAIN` vazio). `GET /healthz` → `{"status":"ok"}`.
6. Front: `API_URL=http://localhost:8080 pnpm build` e `pnpm exec next start -p 3000`.

As chamadas foram feitas com `curl` contra `http://localhost:3000`, sempre com `Origin: http://localhost:3000`, guardando o cookie num arquivo de cookies temporário (apagado ao fim).

## Passos e respostas

### (a) Login pela origem do front

```
POST http://localhost:3000/api/v1/auth/login
Origin: http://localhost:3000
Content-Type: application/json
{"email": "<conta local>", "password": "<senha inicial>"}
```

```
HTTP/1.1 200 OK
content-type: application/json
set-cookie: tj_session=<redigido>; Path=/; Max-Age=28800; HttpOnly; SameSite=Lax
x-request-id: <uuid>
```

- Sem atributo `Domain`: o navegador guarda o cookie como host-only da origem do front. O arquivo de cookies do `curl` registrou o cookie para `localhost` com "incluir subdomínios" = `FALSE`.
- Corpo: `AuthContext` com `must_change_password = true`, papéis `ASSOCIADO, PRESIDENTE`, 30 permissões e `csrf_token` presente.

Troca obrigatória de senha, também pelo rewrite (escrita com `Origin` e `X-CSRF-Token`):

```
POST http://localhost:3000/api/v1/auth/password   → HTTP/1.1 204 No Content
```

Novo login com a senha trocada: `200`, mesmo formato de `Set-Cookie` (sem `Domain`), `must_change_password = false`.

### (b) Escrita autenticada passa pelo `Origin`

```
POST http://localhost:3000/api/v1/financeiro/contas
Origin: http://localhost:3000
Content-Type: application/json
X-CSRF-Token: <token da sessão>
Cookie: tj_session=<redigido>
{"tipo":"DESPESA","nome":"Conta de teste F1"}
```

```
HTTP/1.1 201 Created
content-type: application/json
{"ativo":true,"created_at":"…","id":"<uuid>","nome":"Conta de teste F1","parent_id":null,"tipo":"DESPESA"}
```

Controles negativos, pela mesma sessão e pelo mesmo rewrite (mostram que `Origin` e `X-CSRF-Token` chegam à API e são verificados por ela):

| Requisição | Resposta |
| --- | --- |
| mesma escrita com `Origin: http://evil.example` | `403`, `code: origin_not_allowed` |
| mesma escrita sem `X-CSRF-Token` | `403`, `code: csrf_invalid` |

Um lançamento foi criado para o upload: `POST /api/v1/financeiro/lancamentos` (`DESPESA`, 12345 centavos, `PIX`) → `201 Created`, `status: CRIADA`.

### (c) Upload de comprovante de 9,5 MiB

Arquivo gerado localmente: PDF válido para a detecção de tipo (cabeçalho `%PDF-1.4`, um objeto `Catalog`, preenchimento e `%%EOF`), **9961472 bytes** (9,5 MiB), SHA-256 `4CE589DE1559A239C54D0E01B765B507844FB7A7E82C92C9DAD4D643DD9D809D`.

```
POST http://localhost:3000/api/v1/financeiro/lancamentos/<id>/comprovantes
Origin: http://localhost:3000
X-CSRF-Token: <token da sessão>
Content-Type: multipart/form-data; boundary=…   (campo file, type=application/pdf)
```

```
HTTP/1.1 100 Continue
HTTP/1.1 201 Created
content-type: application/json
{"content_type":"application/pdf","id":"<uuid>","original_filename":"comprovante-9_5MiB.pdf","size_bytes":9961472,"uploaded_at":"…","uploaded_by":"<uuid>","version":1}
```

- Corpo enviado (multipart): 9961689 bytes.
- `size_bytes` = 9961472 = tamanho do arquivo.
- Integridade: `GET /api/v1/financeiro/comprovantes/<id>/url` pelo `:3000` → `200`; o download pela URL assinada → `200`, 9961472 bytes, SHA-256 idêntico ao do arquivo enviado.

## `API_URL` no build e na execução

| Comando | Sem `API_URL` | Observação |
| --- | --- | --- |
| `next build` | **falha** (sai com 1): `Build error occurred` / `Error: API_URL não está definida: o rewrite de /api/v1/* precisa do endereço da API (por exemplo API_URL=http://localhost:8080).` | o rewrite é calculado no build |
| `next typegen` | **falha** (sai com 1): `Unexpected error while generating route types` com a mesma mensagem | o gate Web roda `next typegen` antes do `tsc`; sem ele, o `tsc` falha em `LayoutProps` |
| `next dev` | **falha** ao subir, com a mesma mensagem | o rewrite é calculado ao carregar a configuração |
| `next start` | sobe normalmente | usa o destino gravado no build |

O destino fica **fixado no build**: o build com `API_URL=http://localhost:8080` grava `"destination": "http://localhost:8080/api/v1/:path*"` em `.next/routes-manifest.json`. Depois, `next start` sem `API_URL`, e também com `API_URL=http://localhost:9`, continuou encaminhando para `:8080` (`GET /api/v1/auth/me` → `401 unauthenticated` vindo da API). Mudar o endereço da API exige novo build.

Consequência para os gates: o gate Web (`pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`) precisa de `API_URL` no ambiente. F1 o rodou com `API_URL=http://localhost:8080`. O CI precisa da mesma variável nos passos `next typegen` e `next build`; F1 não altera `.github/`.

## Limite de corpo de 10 MB no rewrite (achado)

O ADR-010 e o WEB-D-009 contam com que só o `proxy.ts` limite o corpo a 10 MB. Na prática, **o rewrite para destino externo também limita**, mesmo sem `proxy.ts` no projeto, apesar de a documentação de `proxyClientMaxBodySize` dizer que ele só vale "when proxy is used".

| Arquivo enviado pelo `:3000` | Corpo multipart | Resposta ao navegador |
| --- | --- | --- |
| 9,5 MiB (9961472 bytes) | 9961689 bytes | `201`, íntegro (item c) |
| 10 MiB − 1 KiB (10484736 bytes) | 10484957 bytes | `201`, `size_bytes` 10484736 |
| 10,5 MiB (11010048 bytes) | 11010266 bytes | **`500`, corpo `Internal Server Error` (texto, não `problem+json`)** |

No caso de 10,5 MiB, o log do `next start` registrou `Request body exceeded 10MB for /api/v1/financeiro/lancamentos/<id>/comprovantes. Only the first 10MB will be available unless configured.` seguido de `Failed to proxy http://localhost:8080/... Error: socket hang up` (`ECONNRESET`). A API recebeu o corpo truncado e respondeu `422` depois de cerca de 30 s, resposta que não chegou ao navegador.

Impacto:

- Arquivos até cerca de 10 MiB menos o cabeçalho multipart (por volta de 200 bytes) passam. Um arquivo dentro do limite da API (10 MiB) mas cujo corpo passe de 10 MB (10485760 bytes) falha nessa faixa estreita.
- Um arquivo acima do limite recebe `500` em texto em vez de `413 document_too_large`: a mensagem de FND-06 AC4 ("grande demais") não aparece; a tela mostraria a mensagem genérica de erro HTTP 500.

Verificação exploratória, **não commitada**: com `experimental: { proxyClientMaxBodySize: "12mb" }` no `next.config.ts` (novo build), o mesmo upload de 10,5 MiB recebeu `413` `application/problem+json` com `code: document_too_large`, vindo da API. A alteração foi revertida; o `next.config.ts` commitado não tem essa opção.

Decisão pendente do mantenedor (fora da autoridade de F1, que não contorna no front): aceitar o limite como está (com verificação de tamanho no cliente em FIN-b), adotar a opção experimental acima, ou a opção (b) do ADR-010.

## Desmontagem

API e `next start` encerrados, contêiner Garage removido (`docker rm -f`), `docker compose down -v` (rede e volume `f1_pgdata` removidos), `.env` local, arquivo de cookies, senhas temporárias e PDFs gerados apagados. Nenhum processo da verificação ficou em execução.
