# Smoke manual da Web V1 contra a stack real (INT-03)

- **Data**: 2026-10-04
- **Unidade**: INT (tasks `06-integracao.md`, T3), branch `feature/web-int2`
- **Código exercitado**: commit `6926cb4` (base `ccc372c`, com a navegação de T1 em `e6cc116`)
- **Versões**: Next.js 16.3.8, Node 22.18.0, Go 1.26.1, PostgreSQL 16 (compose), Garage `dxflrs/garage:v2.4.1`

## Leia antes

**Atualização de 2026-10-09:** a interação visual descrita abaixo como pendente foi exercitada depois, num navegador real, e está registrada em `smoke-visual.md`. Ela encontrou um defeito funcional (`/financeiro/contas` não abria no build de produção), corrigido no commit `5514b38`.

> **A interação visual no navegador não foi exercitada.** O agente da INT não tinha navegador automatizado (Playwright está fora da V1, WEB-D-017, e nenhuma ferramenta foi instalada). Cliques, renderização das telas, mensagens exibidas, diálogos de confirmação e redirecionamentos feitos pelo JavaScript **não** foram vistos por ninguém nesta execução. Esses comportamentos estão cobertos pelos testes de componente com MSW de cada sub-spec, mas a confirmação na tela fica **pendente de execução humana**, pelo roteiro da seção "Roteiro para o mantenedor".
>
> **Redefinição de senha com token válido: não executável localmente** (INT-03 AC3). O provedor de e-mail `log` registra só o domínio do destinatário e o assunto (`"e-mail simulado (provedor log)"`, `recipient_domain`, `subject`); o link com o token nunca aparece. O pedido de recuperação e o token inválido foram executados (itens 3 e 4). A redefinição com token válido fica coberta pelos testes com MSW de ACS-03 (`web/features/identity/acesso/`, `web/app/(public)/redefinir-senha/page.test.tsx`).

Este arquivo não contém senha, token, cookie, chave S3 nem dado pessoal. As contas usadas são fictícias (`*@local.test`), as senhas foram geradas na hora, guardadas só em arquivos temporários fora do repositório e apagadas ao fim. Valores de `Set-Cookie` aparecem como `<redigido>`.

## Como foi executado

Stack montada a partir do worktree da INT, seguindo `docs/development/web.md`:

1. Banco: `docker compose up -d db` e `docker compose --profile tools run --rm migrate` (migrações 1 a 9).
2. S3: contêiner Garage avulso na porta `3900`, com o `garage.toml` de `api/internal/platform/testutil/s3.go`, `rpc_secret` e chaves geradas na hora, bucket `tj-local`.
3. Primeira conta: `go run ./cmd/bootstrap-admin --role PRESIDENTE` (senha só em `BOOTSTRAP_ADMIN_PASSWORD`). Uma primeira tentativa com o padrão `ADMIN_SISTEMA` mostrou que essa conta não consegue promover ninguém aos papéis operacionais (ver "Achados"); o banco foi recriado (`docker compose down -v`) e a execução registrada abaixo é toda com `PRESIDENTE`.
4. API: binário de `./cmd/api` com `APP_ENV=development`, `COOKIE_SECURE=false`, `ALLOWED_ORIGINS=http://localhost:3000`, `APP_BASE_URL=http://localhost:3000`, `EMAIL_PROVIDER=log`, `STORAGE_ENABLED=true` e `S3_*` do Garage. `.env` local copiado do exemplo, ignorado pelo git e apagado ao fim.
5. Front: `API_URL=http://localhost:8080 pnpm build` e `pnpm exec next start -p 3000`.

Forma de execução de cada item:

- **(A) chamadas da tela**: `curl` contra `http://localhost:3000` (pelo rewrite, nunca direto no `:8080`), com `Origin: http://localhost:3000`, cookie guardado num arquivo de cookies por conta e `X-CSRF-Token` (vindo de `GET /api/v1/auth/me`) em toda escrita, como o navegador faz. Corpo e caminho iguais aos montados pelos hooks de `web/features/**`. Conferidos status, `code` de `problem+json` e campos do corpo.
- **(H) HTML da rota**: `GET` da rota no `next start`, conferindo status e `<title>`.
- **Papéis reais**: contas de teste criadas e promovidas pela própria API de administração, com o `PRESIDENTE` (`TESOURARIA`, `ESTOQUE_LOJA`, `CONSELHO_FISCAL`, `ADMIN_SISTEMA` e uma conta só `ASSOCIADO`). As permissões efetivas devolvidas por `GET /auth/me` para cada uma foram iguais, uma a uma, às usadas nos testes de T1 (`web/app/(app)/inicio/page.test.tsx`).

## Resultados

| # | Item (INT-03 AC2) | Resultado | Como | Observação |
| --- | --- | --- | --- | --- |
| 1 | Login | Passou | A, H | `POST /auth/login` → `200`, `Set-Cookie: tj_session=<redigido>; Path=/; Max-Age=28800; HttpOnly; SameSite=Lax`, sem `Domain` (host-only na origem do front). Corpo: `must_change_password=true`, papéis `ASSOCIADO, PRESIDENTE`, 30 permissões, `csrf_token` presente. Senha errada → `401 invalid_credentials`. `/entrar` → `200`, título "Entrar \| TJ Platform". |
| 2 | Troca obrigatória de senha | Passou | A, H | Antes da troca, `GET /financeiro/contas` → `403 password_change_required`. `POST /auth/password` sem `X-CSRF-Token` → `403 csrf_invalid`; com ele → `204`. Depois, `GET /auth/me` → `must_change_password=false`; novo login com a senha nova → `200`. Repetido para as cinco contas de teste no primeiro acesso: todas `204`. `/conta/senha` → `200`, "Trocar senha". |
| 3 | Pedido de recuperação (mensagem neutra) | Passou | A, H | `POST /auth/password-reset/request` com conta existente e com e-mail inexistente → ambos `202`, corpo idêntico (`"Se existir uma conta vinculada ao e-mail informado, enviaremos instruções."`). O log da API registrou só o domínio e o assunto. `/recuperar-acesso` → `200`, "Recuperar acesso". |
| 4 | `/redefinir-senha` com token inválido | Passou | A, H | `POST /auth/password-reset/confirm` com token inexistente → `400 invalid_reset_token` ("O link de recuperação é inválido ou expirou."). `/redefinir-senha` → `200`, "Redefinir senha". O token na URL (fragmento `#token=`) e sua remoção da barra de endereço só são observáveis no navegador: roteiro humano. |
| 4b | `/redefinir-senha` com token válido | Não executável localmente | — | Provedor `log` não expõe o link (INT-03 AC3). Coberto pelos testes com MSW de ACS-03. |
| 5 | Criação e promoção de usuário | Passou | A, H | Com `PRESIDENTE`: `POST /users` → `201` (papéis `ASSOCIADO`, `must_change_password=true`) para cinco contas; `POST /users/{id}/admin-membership` com `TESOURARIA`, `ESTOQUE_LOJA`, `CONSELHO_FISCAL` e `ADMIN_SISTEMA` → `204`, e o `GET /auth/me` de cada conta passou a trazer o papel. Ajuste de papéis (`PUT /users/{id}/roles`, acrescentar e retirar `DIRETORIA`) → `204`. E-mail repetido → `409 email_taken`; `Origin` externa → `403 origin_not_allowed`; sem `X-CSRF-Token` → `403 csrf_invalid`. Listagem com `role=TESOURARIA` → 1 item; `limit=2` → 2 itens e `next_cursor`. Desativar e reativar → `204`; retirar acesso de quem não é administrativo → `409 not_admin`; retirar acesso do `ADMIN_SISTEMA` de teste → `204`, e a sessão aberta dele passou a receber `401`. `ADMIN_SISTEMA` tentando promover a `TESOURARIA` → `403 privilege_escalation` (regra da API; ver "Achados"). `/admin/usuarios` → `200`, "Usuários". |
| 6 | Senha temporária | Passou | A | `POST /users/{id}/password-reset` sem motivo → `422 reason_required`; com motivo → `200`, `temporary_password` presente e `Cache-Control: no-store`. Login com a senha temporária → `200` com `must_change_password=true`; troca → `204`. `ADMIN_SISTEMA` também gerou senha temporária para uma conta só `ASSOCIADO` → `200`. A exibição única na tela e o descarte do valor ficam no roteiro humano. |
| 7 | Ciclo completo do financeiro com comprovante | Passou | A, H | Como `TESOURARIA`: contas `RECEITA`, `DESPESA` e subconta → `201`; renomear → `200`; desativar → `204` (lista: 3 contas, 2 ativas). Lançamento `RECEITA` de 15000 centavos → `201 CRIADA`; edição para 20000 com taxa 150 e `CARTAO` → `200`, líquido 19850; receber → `204`; receber de novo → `409 lancamento_nao_pode_ser_recebido`. Devolução de 5000 sobre a receita → `201` (`DESPESA`, `CRIADA`, `devolucao_de_id` certo); pagar → `204`. `DESPESA` de 3000 → `201`; cancelar com motivo só de espaços → `422 motivo_obrigatorio`; com motivo → `204` (`CANCELADA`, motivo gravado). Saldo → `200`, `saldo_cents=14850` (19850 − 5000; cancelado fora). Comprovante PDF de 200052 bytes → `201` (`size_bytes` igual, `version=1`); sem `X-CSRF-Token` → `403`; listar → 1 item; URL assinada → `200` e download com SHA-256 igual ao enviado. Comprovante exatamente no limite do cliente (10.420.224 bytes) pelo rewrite → `201`, `size_bytes=10420224`, download íntegro. `/financeiro/contas`, `/financeiro/lancamentos`, `/financeiro/lancamentos/{id}` e `/financeiro/saldo` → `200`. |
| 8 | Ciclo completo do estoque com ajuste negativo | Passou | A, H | Como `ESTOQUE_LOJA`: produto → `201`; código repetido → `409 codigo_duplicado`. Entrada 10, saída 4 e devolução 1 (da saída), todas com origem `INVENTARIO` → `201`. Saída de 100 → `409 saldo_insuficiente`. Ajuste −2 sem motivo → `422 motivo_obrigatorio`; com motivo → `201` (`AJUSTE`, origem `AJUSTE_MANUAL`, quantidade −2). Movimentações listadas na ordem, com o motivo do ajuste. Saldo → `5` (10 − 4 + 1 − 2). `/estoque/produtos` e `/estoque/produtos/{id}` → `200`. |
| 9 | Visão de `CONSELHO_FISCAL` sem ações de escrita | Passou (API e permissões); tela pendente | A | Leituras → `200`: contas, lançamentos, saldo, comprovantes, URL do comprovante, produtos, movimentações e saldo de estoque. Todas as escritas → `403 forbidden`: criar, renomear e desativar conta; criar, pagar e cancelar lançamento; anexar comprovante; criar produto; movimentação; ajuste. `GET /users` → `403`. `GET /auth/me` traz só permissões de leitura (nenhuma de escrita), então pelos testes de cada tela nenhuma ação de escrita é renderizada; ver isso na tela fica no roteiro humano. Cruzados: `TESOURARIA` em estoque, `ESTOQUE_LOJA` em financeiro e `ASSOCIADO` em financeiro → `403 forbidden`. |
| 10 | Sessão expirada (cookie removido) | Passou | A | Sem o cookie: `GET /auth/me`, `GET /financeiro/lancamentos` e `POST /financeiro/contas` → `401 unauthenticated`. Expiração real: API reiniciada com `SESSION_IDLE_MINUTES=1`, login, leitura `200`, 70 s ociosa, leitura → `401 session_expired` ("A sessão expirou. Entre novamente."); a API voltou ao padrão depois. O redirecionamento para `/entrar?sessao=encerrada` é feito pelo front e fica no roteiro humano. |
| 11 | Logout | Passou | A | Sem `X-CSRF-Token` → `403 csrf_invalid`. Com ele → `204` e `Set-Cookie: tj_session=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax`. Depois, `GET /auth/me` → `401`; reusar o cookie antigo → `401 unauthenticated`. |
| 12 | Raiz e páginas da navegação (INT-01) | Passou | H | `/` → `307` para `/inicio`, sem chamar a API (o log da API não registrou `/healthz` ao abrir `/`). `/inicio` → `200`, "Início"; `/sem-acesso` → `200`, "Sem acesso". As permissões reais das cinco contas batem com as dos testes de T1, que conferem itens de navegação e atalhos por perfil. |

Nenhum item falhou. Nenhuma correção foi aberta em frente dona.

## Achados

1. **`ADMIN_SISTEMA` não promove aos papéis operacionais.** A API recusa conceder papel com permissão que o ator não tem (`403 privilege_escalation`, regra já especificada em `fundacao-core`). Um `ADMIN_SISTEMA` só tem permissões de `identity` e `audit`, então não promove ninguém a `TESOURARIA`, `ESTOQUE_LOJA` nem `CONSELHO_FISCAL`; só um `PRESIDENTE` consegue. A Web mostra a mensagem certa ("Você não pode conceder ou alterar permissões que não tem.", catálogo de ID-ADM). Não é defeito da Web. É inconsistência documental: o item 2 da definição de "pronta para uso" em `.specs/features/web/STATE.md` diz "Um administrador (`ADMIN_SISTEMA` ou `PRESIDENTE`) consegue ... promovê-la ... com os papéis certos (por exemplo, `TESOURARIA` ou `ESTOQUE_LOJA`)", o que só vale para `PRESIDENTE`. Registrado para a rodada de documentação; o guia (`docs/development/web.md`) já orienta a criar a primeira conta como `PRESIDENTE`.
2. **Motivos administrativos.** A promoção recusou o motivo `"Smoke"` (5 caracteres) com `422 reason_required` (mínimo de 10 caracteres na API). Na retirada de acesso, um motivo de 5 caracteres foi aceito (`204`). Comportamento da API, fora da Web; a Web mostra a mensagem de `reason_required` quando a API a devolve.
3. **`413` da API pelo rewrite, intermitente.** Uma vez, um corpo JSON acima do limite da API (gerado por engano no próprio script do smoke, nunca pela tela) chegou ao cliente como `500`: o Next registrou `ECONNRESET` ao encaminhar, porque a API respondeu `413` e fechou a conexão antes de ler o corpo inteiro. Repetido de propósito com um corpo de cerca de 200 KB, chegou como `413 payload_too_large`. As telas não montam corpos JSON desse tamanho (motivos têm no máximo 500 caracteres); o caso dos uploads é o já tratado pela checagem de 10.420.224 bytes do cliente (WEB-D-009).

## Roteiro para o mantenedor (navegador)

Suba a stack pelo `docs/development/web.md` (primeira conta como `PRESIDENTE`) e abra `http://localhost:3000` num navegador comum. Para cada passo, anote passou/falhou e data nesta seção.

1. `/` leva a `/entrar`. Entrar com a senha inicial leva à troca obrigatória (`/conta/senha`); o menu mostra só "Sair". Trocar a senha leva a `/inicio`.
2. Em `/inicio`, conferir os atalhos e a navegação lateral: `PRESIDENTE` vê Financeiro (Lançamentos, Contas, Saldo), Estoque (Produtos) e Administração (Usuários). Estreitar a janela: a navegação vira menu recolhível.
3. `/recuperar-acesso`: enviar um e-mail qualquer mostra a mensagem neutra. Abrir `/redefinir-senha#token=qualquer`: o token some da barra de endereço; enviar uma senha nova mostra a mensagem de link inválido ou expirado.
4. `/admin/usuarios`: criar uma pessoa, promovê-la a `TESOURARIA` com motivo de 10+ caracteres, editar papéis, gerar a senha temporária (aparece uma vez; fechar o diálogo e reabrir não a mostra de novo), desativar e reativar.
5. Entrar como a tesouraria: fazer a troca obrigatória; criar contas, um lançamento com valor em reais (por exemplo `1.234,56`), editar, receber, registrar devolução, pagar, cancelar com motivo (o botão fica desabilitado com motivo vazio) e conferir o saldo. Anexar um PDF pequeno, listar e baixar. Tentar anexar um arquivo acima de 10 MiB: a tela recusa antes de enviar.
6. Promover outra pessoa a `ESTOQUE_LOJA` e entrar com ela: criar produto, entrada, saída, devolução, ajuste negativo com motivo (confirmação obrigatória) e conferir o saldo no detalhe do produto. A entrada, a saída e a devolução oferecem só a origem `INVENTARIO`.
7. Promover outra pessoa a `CONSELHO_FISCAL` e entrar com ela: as telas de financeiro e estoque abrem só para leitura, sem nenhum botão de criar, editar, receber, pagar, cancelar, anexar, movimentar ou ajustar; `/admin/usuarios` mostra "Sem acesso".
8. Entrar com uma conta só `ASSOCIADO`: `/inicio` mostra "Sua conta ainda não tem acesso a nenhum módulo." e a navegação só tem "Início".
9. Com uma sessão aberta, apagar o cookie `tj_session` nas ferramentas do navegador e clicar em qualquer item: a tela leva a `/entrar` com o aviso de sessão encerrada e, depois do login, volta à página pedida.
10. "Sair" no menu da pessoa leva a `/entrar`; o botão voltar do navegador não mostra dados da sessão anterior.

## Desmontagem

API e `next start` encerrados, contêiner Garage removido (`docker rm -f`), `docker compose down -v` (volume e rede do worktree removidos), `.env` local, arquivos de cookies, senhas geradas, PDFs e configuração do Garage apagados. Nenhum processo do smoke ficou em execução.
