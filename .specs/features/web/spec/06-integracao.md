# Integração e Verificação da Web V1 — Specification

## Problem Statement

As frentes de módulo (sub-specs 02 a 05) são feitas em paralelo e, por regra de propriedade, nenhuma delas monta a navegação global, a página inicial nem o redirecionamento de `/` (WEB-D-013). Sem a integração, as telas existem mas não são encontráveis, e ninguém confere o conjunto contra a stack real. Esta sub-spec liga as frentes, documenta como rodar a Web, executa o smoke manual (WEB-D-006) e entrega o resultado a um verificador independente.

## Goals

- [ ] A Web V1 atende à definição de "pronta para uso" de `.specs/features/web/STATE.md`.
- [ ] Um agente que não escreveu o código confirma isso com evidência.

## Out of Scope

| Feature | Reason |
| --- | --- |
| E2E automatizado (Playwright) | WEB-D-017. |
| Deploy, hospedagem, domínio do cookie | Fora deste ciclo (ADR-005, ADR-010). |
| Corrigir as inconsistências documentais de `STATE.md` | Registradas; corrigidas em rodada própria de documentação, se o mantenedor autorizar. |

---

## Assumptions & Open Questions

| Assumption / decision | Chosen default | Rationale | Confirmed? |
| --- | --- | --- | --- |
| Página inicial autenticada | `/inicio` com atalhos para as áreas que a pessoa pode ver, gerados da mesma lista da navegação; sem indicadores | Dashboard fora da V1 | y |
| Raiz `/` | Redireciona para `/inicio`; o portão de F2 leva quem não tem sessão a `/entrar` | WEB-D-013 | y |
| Ordem dos itens de navegação | Início, Financeiro (Lançamentos, Contas, Saldo), Estoque (Produtos), Administração (Usuários) | Frequência de uso esperada da tesouraria e do estoque | y |
| Onde fica o checklist de smoke | `.specs/features/web/smoke.md`, com resultado por item e data | Mesmo local das demais evidências da feature | y |
| Guia de desenvolvimento | `docs/development/web.md` | `docs/development/` já reúne os guias de ambiente | y |

**Open questions:** none.

---

## User Stories

### INT-01: Navegação global e página inicial

**User Story**: Como pessoa autenticada, quero encontrar todas as áreas que posso usar a partir de qualquer tela.

**Acceptance Criteria**:

1. The registro `components/app/nav-items.ts` SHALL juntar os itens de `features/identity/nav.ts`, `features/financeiro/nav.ts` e `features/estoque/nav.ts`, mais "Início", na ordem definida, sem redefinir caminho ou permissão de nenhum item.
2. WHEN o shell é renderizado THEN ele SHALL usar `nav-items.ts`, exibindo só os itens cuja permissão está no contexto de sessão.
3. WHEN a pessoa abre `/` THEN o sistema SHALL levá-la a `/inicio`.
4. WHEN a pessoa abre `/inicio` THEN o sistema SHALL mostrar um atalho para cada área visível na navegação e, sem nenhuma área, uma mensagem explicando que a conta ainda não tem acesso a módulos.
5. The página raiz SHALL deixar de consultar `/healthz` no servidor.

**Independent Test**: com MSW, sessões de `TESOURARIA`, `ESTOQUE_LOJA`, `CONSELHO_FISCAL`, `ADMIN_SISTEMA` e `ASSOCIADO`, conferindo itens e atalhos.

---

### INT-02: Guia de desenvolvimento da Web

**User Story**: Como pessoa desenvolvedora, quero saber como rodar o front junto com a API.

**Acceptance Criteria**:

1. The guia `docs/development/web.md` SHALL explicar variáveis (`API_URL`, `ALLOWED_ORIGINS`, `APP_BASE_URL`, `COOKIE_SECURE`, `STORAGE_ENABLED` e `S3_*`), ordem de subida (banco, migrations, armazenamento S3 local, API, front), o rewrite same-origin e os comandos de gate da Web.
2. The guia SHALL registrar o comportamento de `API_URL` no build, conforme a evidência de F1.
3. The guia SHALL não conter segredo, senha real, cookie ou dado pessoal.

**Independent Test**: seguir o guia num clone limpo e chegar à tela de login.

---

### INT-03: Smoke manual contra a stack real

**User Story**: Como mantenedor, quero a confirmação de que o conjunto funciona com API, banco e S3 reais.

**Acceptance Criteria**:

1. WHEN o smoke é executado THEN cada item do checklist SHALL ter resultado (passou ou falhou), data e observação registrados em `.specs/features/web/smoke.md`.
2. The checklist SHALL cobrir: login, troca obrigatória, pedido de recuperação (mensagem neutra), `/redefinir-senha` com token inválido (`invalid_reset_token`), criação e promoção de usuário, senha temporária, ciclo completo do financeiro com comprovante, ciclo completo do estoque com ajuste negativo, visão de `CONSELHO_FISCAL` sem ações de escrita, sessão expirada (cookie removido) e logout.
3. The registro SHALL declarar que a redefinição com token válido não é executável na stack local, porque o provedor de e-mail `log` registra só o domínio do destinatário e o assunto; esse caminho fica coberto pelos testes com MSW de ACS-03.
4. IF algum item falhar THEN a INT SHALL registrar a falha, abrir a correção na frente dona do arquivo e repetir o item, sem marcar a Web como pronta.
5. The registro SHALL não conter senha, token, cookie ou dado pessoal real.

**Independent Test**: o próprio checklist.

---

### INT-04: Verificação independente

**User Story**: Como mantenedor, quero que um agente que não escreveu o código confirme a Web V1.

**Acceptance Criteria**:

1. WHEN todas as tasks das sub-specs 01 a 06 (até INT-03) estão concluídas THEN um agente novo, sem o contexto dos autores, SHALL verificar cada AC das sub-specs 01 a 05 e INT-01 a INT-03 contra os testes, citando `arquivo:linha` como evidência.
2. The verificador SHALL rodar o sensor de discriminação: injetar falhas de comportamento numa cópia isolada (worktree temporário), confirmar que os testes as matam e confirmar que a árvore real ficou igual ao estado anterior ao sensor.
3. The sensor SHALL incluir, no mínimo: remover o `X-CSRF-Token` das escritas; não redirecionar em `401`; exibir uma ação sem a permissão; aceitar `next` externo; enviar dinheiro como número com casas decimais; enviar cancelamento ou ajuste com motivo vazio; pular a confirmação de uma ação sensível; oferecer origem reservada em movimentação manual; guardar a senha temporária no cache; manter o token de recuperação na URL.
4. The verificador SHALL escrever `.specs/features/web/validation.md` com veredito (PASS ou FAIL), evidência por AC, resultado do sensor e intervalo de commits verificado.
5. IF o veredito for FAIL THEN cada lacuna SHALL virar task de correção na frente dona do arquivo, com no máximo 3 ciclos de correção e nova verificação antes de levar ao mantenedor.

**Independent Test**: `validate_state.py web` passa com o `validation.md` produzido.

---

## Edge Cases

- WHEN um módulo não exporta nenhum item visível para a sessão THEN a seção daquele módulo SHALL não aparecer na navegação.

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
| --- | --- | --- | --- |
| INT-01 | INT-01: Navegação global e página inicial | Tasks | In Tasks |
| INT-02 | INT-02: Guia de desenvolvimento | Tasks | In Tasks |
| INT-03 | INT-03: Smoke manual | Tasks | In Tasks |
| INT-04 | INT-04: Verificação independente | Tasks | In Tasks |

**Coverage:** 4 total, 4 mapped to tasks, 0 unmapped.

---

## Success Criteria

- [ ] `validation.md` com PASS e evidência; smoke registrado sem falha pendente.
- [ ] Todos os gates da Web verdes, backend intacto.
