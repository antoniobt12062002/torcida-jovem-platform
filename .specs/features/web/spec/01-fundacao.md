# Fundação Web — Specification

## Problem Statement

O `web/` de hoje é o esqueleto do `create-next-app`: dois componentes de UI (`Button` e `Badge`), `lib/money.ts` e uma página que só consulta `/healthz`. Não há acesso do navegador à API, sessão, tratamento de erro nem componentes para formulários. Nenhuma tela de módulo pode ser feita sem isso. Esta sub-spec entrega a base comum: o caminho same-origin até a API (ADR-010), o cliente tipado, a sessão com CSRF, a proteção de rotas, o kit de UI e os componentes de aplicação que as frentes de módulo usam.

## Goals

- [ ] O navegador fala com a API só pela origem do front, e o modelo de segurança da API (cookie, `Origin`, CSRF) funciona sem nenhuma mudança no backend.
- [ ] As frentes de módulo (sub-specs 02 a 05) conseguem ser feitas em paralelo, sem alterar nenhum arquivo desta fundação.

## Out of Scope

| Feature | Reason |
| --- | --- |
| Telas de módulo (senha, usuários, financeiro, estoque) | Sub-specs 02 a 05. |
| Navegação global montada e página inicial | Sub-spec 06 (INT), WEB-D-013. |
| `proxy.ts` para proteger rotas | WEB-D-009: trunca uploads acima de 10 MB. |
| Bibliotecas além das de WEB-D-003 | WEB-D-003 e WEB-D-011. |

---

## Assumptions & Open Questions

| Assumption / decision | Chosen default | Rationale | Confirmed? |
| --- | --- | --- | --- |
| Como o navegador chega à API | Rewrite `/api/v1/*` → `API_URL` no `next.config.ts` | WEB-D-001, ADR-010 | y |
| Onde fica o token CSRF | Em memória, no provedor de sessão, vindo de `POST /auth/login` ou `GET /auth/me`; nunca em `localStorage`, `sessionStorage` ou cookie legível | AD-008/ADR-005: o segredo de sessão nunca fica exposto; o token CSRF é recuperável a qualquer momento por `/auth/me` | y |
| Validação de formulário | Regras do React Hook Form mais mapeamento dos `errors[]` de `problem+json` para os campos; sem biblioteca de schema | WEB-D-003 | y |
| Proteção de rotas | Layout autenticado (Client Component) consulta `/auth/me`; API continua sendo a autoridade | WEB-D-009 | y |
| Reação a `403 csrf_invalid` | Rebuscar `/auth/me` e pedir que a pessoa repita a ação; sem repetição automática da escrita | A API recusou a escrita antes de executá-la, mas repetir automaticamente uma escrita é arriscado e não traz ganho para a V1 | y |
| Toast | `Toast` do `@base-ui/react` | WEB-D-003 | y |

**Open questions:** none.

---

## User Stories

### FND-01: Acesso same-origin à API

**User Story**: Como front, quero chamar a API pela minha própria origem, com o cliente tipado pelo contrato, para usar o modelo de segurança da API sem mudá-lo.

**Acceptance Criteria**:

1. WHEN o navegador requisita qualquer caminho sob `/api/v1/` na origem do front THEN o servidor Next SHALL encaminhar a requisição para o mesmo caminho em `API_URL`, preservando método, corpo, cabeçalhos `Cookie`, `Origin`, `Content-Type` e `X-CSRF-Token`, e devolver ao navegador o status, o corpo e os cabeçalhos `Set-Cookie`, `Content-Type`, `Retry-After` e `Cache-Control` da API.
2. The cliente da API SHALL ser criado com `openapi-fetch` sobre os tipos gerados em `web/lib/api/*.d.ts`, com `credentials: "same-origin"`, e nenhuma tela SHALL chamar `fetch` diretamente para a API.
3. WHEN o cliente envia uma requisição com método diferente de `GET`, `HEAD` ou `OPTIONS` THEN ele SHALL incluir o cabeçalho `X-CSRF-Token` com o token da sessão atual.
4. WHEN a API responde com `application/problem+json` THEN o cliente SHALL entregar à tela um erro tipado com `status`, `code` e `errors[]` (campo e código), sem lançar exceção genérica.
5. WHEN a verificação de aceite contra a API real é executada THEN ela SHALL demonstrar que (a) o cookie `tj_session` chega ao navegador como cookie da origem do front, (b) uma escrita autenticada passa pelo `Origin` sem `403 origin_not_allowed`, e (c) um comprovante de pelo menos 9,5 MiB chega íntegro à API pelo rewrite.
6. IF qualquer item do AC 5 falhar THEN a unidade F1 SHALL parar e reportar, sem contorno no front nem mudança na API.

**Independent Test**: com a stack local, fazer login pela origem `http://localhost:3000`, criar um lançamento e anexar um PDF de 9,5 MiB, tudo pelo rewrite.

---

### FND-02: Kit de UI

**User Story**: Como frente de módulo, quero os componentes de interface já instalados num só lugar, para montar telas sem instalar nada.

**Acceptance Criteria**:

1. The kit SHALL oferecer, em `web/components/ui/`, os componentes de campo de texto, rótulo, agrupamento de campo com mensagem de erro, seleção, área de texto, caixa de seleção, diálogo, diálogo de confirmação, tabela, menu, esqueleto de carregamento, alerta e toast, além dos já existentes `Button` e `Badge`.
2. WHEN um componente de campo recebe uma mensagem de erro THEN ele SHALL associá-la ao campo por `aria-describedby` e marcar o campo com `aria-invalid`.
3. IF instalar um componente exigir uma dependência que ainda não está em `package.json` THEN a unidade F4 SHALL parar e reportar, sem alterar `package.json`.

**Independent Test**: renderizar cada componente num teste e confirmar o papel acessível e a associação de erro.

---

### FND-03: Sessão, login e logout

**User Story**: Como pessoa com conta, quero entrar e sair, e quero que o sistema perceba quando a minha sessão acabou.

**Acceptance Criteria**:

1. WHEN a pessoa envia e-mail e senha válidos em `/entrar` THEN o sistema SHALL chamar `POST /api/v1/auth/login`, guardar o contexto de sessão devolvido e levar a pessoa ao destino pedido (parâmetro `next`, aceito só se for caminho interno começando com `/` e sem `//`) ou à página inicial.
2. IF a API responde `401 invalid_credentials` THEN o sistema SHALL mostrar "E-mail ou senha inválidos." sem indicar qual dos dois está errado.
3. IF a API responde `429 login_blocked` THEN o sistema SHALL informar que o acesso está temporariamente bloqueado e quanto tempo esperar, a partir de `Retry-After`.
4. WHEN o contexto de sessão tem `must_change_password = true` THEN o sistema SHALL levar a pessoa a `/conta/senha` e não SHALL mostrar outra tela autenticada até a troca.
5. WHEN a pessoa escolhe sair THEN o sistema SHALL chamar `POST /api/v1/auth/logout`, limpar todo o cache de consultas e o contexto de sessão e levar a pessoa a `/entrar`.
6. WHEN qualquer chamada autenticada recebe `401` (`unauthenticated` ou `session_expired`) THEN o sistema SHALL limpar o cache e o contexto de sessão e levar a pessoa a `/entrar`, com aviso de sessão encerrada e o caminho atual em `next`.
7. WHEN qualquer chamada recebe `403 password_change_required` THEN o sistema SHALL levar a pessoa a `/conta/senha`.
8. WHEN uma escrita recebe `403 csrf_invalid` THEN o sistema SHALL rebuscar `GET /api/v1/auth/me` e pedir que a pessoa repita a ação, sem repeti-la automaticamente.
9. The token CSRF e o contexto de sessão SHALL ficar só em memória, nunca em `localStorage`, `sessionStorage` ou cookie criado pelo front.
10. The página `/entrar` SHALL oferecer o link "Esqueci minha senha" para `/recuperar-acesso`.

**Independent Test**: com MSW, simular login, troca obrigatória, sessão expirada no meio do uso e logout, e conferir cada redirecionamento.

---

### FND-04: Proteção de rotas e permissões na interface

**User Story**: Como pessoa autenticada, quero ver só o que posso usar; como sistema, quero que a API continue decidindo.

**Acceptance Criteria**:

1. WHEN alguém sem sessão abre uma rota de `app/(app)/` THEN o sistema SHALL levar a pessoa a `/entrar` com o caminho pedido em `next`, sem mostrar conteúdo da rota.
2. WHILE a sessão está sendo consultada THE sistema SHALL mostrar um estado de carregamento, sem mostrar conteúdo protegido.
3. The front SHALL decidir a exibição de cada ação pelas permissões efetivas (`permissions`) do contexto de sessão, nunca pelo nome do papel.
4. IF a pessoa não tem a permissão de leitura de uma área THEN a rota da área SHALL mostrar a página "Sem acesso", sem chamar a API daquela área.
5. IF a API responde `403 forbidden` a uma ação que a interface exibiu THEN o sistema SHALL mostrar "Você não tem permissão para esta ação." e não SHALL alterar o estado exibido.

**Independent Test**: com MSW, abrir rotas protegidas sem sessão, com sessão sem a permissão e com a permissão, e conferir o que aparece.

---

### FND-05: Estrutura da aplicação e componentes compartilhados

**User Story**: Como frente de módulo, quero o shell, os estados de tela e os campos de dinheiro e quantidade prontos, para que todas as telas se comportem igual.

**Acceptance Criteria**:

1. The shell SHALL mostrar o nome da pessoa, o acesso a "Trocar senha" e "Sair", e a navegação gerada a partir de uma lista de itens (rótulo, caminho, permissão exigida), exibindo só os itens cuja permissão está no contexto de sessão.
2. WHEN a largura da tela é menor que o ponto de quebra de tablet THEN a navegação SHALL virar um menu recolhível, e nenhuma ação SHALL ficar inacessível.
3. The aplicação SHALL oferecer estados padronizados de carregamento, lista vazia e erro com opção de tentar de novo.
4. The campo de dinheiro SHALL aceitar valores no formato pt-BR, convertê-los para centavos inteiros com `parseBRL` e nunca SHALL produzir número de ponto flutuante para a API.
5. The campo de quantidade SHALL aceitar só inteiros, com o sinal permitido configurável (só positivos, ou diferente de zero com sinal).
6. WHEN uma ação irreversível ou sensível é acionada THEN o diálogo de confirmação SHALL exigir uma confirmação explícita antes de chamar a API, e, quando configurado com motivo obrigatório, SHALL manter o botão de confirmar desabilitado enquanto o motivo estiver vazio ou só com espaços.
7. The aplicação SHALL exibir valores em centavos com `formatBRL` e datas no formato pt-BR, no fuso de São Paulo.
8. WHILE o contexto de sessão tem `must_change_password = true` THE shell SHALL esconder a navegação e mostrar só a opção de sair.

**Independent Test**: renderizar o shell com dois conjuntos de permissões, os campos de dinheiro e quantidade com entradas válidas e inválidas, e o diálogo com motivo vazio.

---

### FND-06: Erros da API apresentados à pessoa

**User Story**: Como pessoa usando o sistema, quero entender por que uma ação falhou.

**Acceptance Criteria**:

1. WHEN a API responde `422 validation_failed` com `errors[]` THEN o formulário SHALL mostrar cada erro junto do campo correspondente, quando o campo existir no formulário.
2. WHEN a API responde com um `code` que a tela conhece THEN o sistema SHALL mostrar a mensagem em português definida para aquele `code`.
3. IF o `code` não está no catálogo da tela THEN o sistema SHALL mostrar uma mensagem genérica com o status HTTP, sem expor detalhes internos.
4. IF a API responde `413` THEN o sistema SHALL informar que o conteúdo enviado é grande demais.
5. IF a API está indisponível (`503` ou falha de rede) THEN o sistema SHALL informar que o serviço está indisponível e oferecer tentar de novo.

**Independent Test**: com MSW, devolver cada tipo de erro a um formulário de exemplo nos testes e conferir a mensagem e o campo.

---

## Edge Cases

- WHEN duas chamadas recebem `401` quase ao mesmo tempo THEN o sistema SHALL fazer um único redirecionamento para `/entrar`.
- IF o parâmetro `next` aponta para outra origem (por exemplo `//exemplo.com` ou `https://…`) THEN o sistema SHALL ignorá-lo e usar a página inicial.
- WHEN a pessoa já autenticada abre `/entrar` THEN o sistema SHALL levá-la à página inicial.

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
| --- | --- | --- | --- |
| FND-01 | FND-01: Acesso same-origin à API | Tasks | Verified |
| FND-02 | FND-02: Kit de UI | Tasks | Verified |
| FND-03 | FND-03: Sessão, login e logout | Tasks | Verified |
| FND-04 | FND-04: Proteção de rotas e permissões | Tasks | Verified |
| FND-05 | FND-05: Estrutura e componentes compartilhados | Tasks | Verified |
| FND-06 | FND-06: Erros da API apresentados | Tasks | Verified |

**Coverage:** 6 total, 6 mapped to tasks, 0 unmapped.

---

## Success Criteria

- [ ] Login, sessão expirada, troca obrigatória e logout funcionam contra a API real pelo rewrite.
- [ ] Nenhum arquivo fora de `web/` muda.
- [ ] As frentes de módulo usam o cliente, a sessão, o kit e os componentes de app sem alterá-los.
