# Migração da Web V1 — Specification

## Problem Statement

As telas da Web V1 estão corretas no comportamento e verificadas, mas foram feitas antes do Design System: cada página monta o próprio contêiner e o próprio título, as tabelas só rolam na horizontal, os filtros vivem em memória e os status dependem só de cor. Esta sub-spec leva cada tela existente aos tokens, componentes e padrões das sub-specs 01 a 04, sem mudar o que ela faz, e define como a migração é verificada.

## Goals

- [ ] Todas as telas da Web V1 usam somente os componentes e padrões do Design System.
- [ ] Nenhum critério de aceite da Web V1 deixa de valer; a API e os contratos não mudam.
- [ ] A verificação cobre acessibilidade, responsividade e aparência num navegador real, não só em jsdom.

## Out of Scope

| Feature | Reason |
| --- | --- |
| Funcionalidade nova, tela nova de módulo, dashboard | WDS-D-010. |
| Mudança de API, contrato OpenAPI, migration ou regra de negócio | A feature não altera o backend. |
| Dependência nova | WEB-D-003. |
| Correção das observações de produto do smoke visual (conta inativa com "Nova subconta", lançamento cancelado com "Anexar") | Regra de negócio dos módulos. |
| E2E automatizado com Playwright | WEB-D-017 continua valendo. |

---

## Assumptions & Open Questions

| Assumption / decision | Chosen default | Rationale | Confirmed? |
| --- | --- | --- | --- |
| Portão de início | A migração só começa com o espécime aprovado | DSF-06 AC6 | y |
| Testes existentes | Continuam passando; mudam só onde uma decisão WDS-D altera texto, estrutura ou ponto de quebra, com a decisão citada | A verificação da Web V1 não pode ser enfraquecida | n |
| Prioridade de colunas por tela | Proposta na tabela "Telas" abaixo, a fechar no Design | Depende do espécime e do uso real | n |
| Árvore do plano de contas no celular | Lista com recuo por nível, sem virar cartões | A hierarquia é a informação principal | n |
| Larguras em que a conferência visual é feita | 390, 768, 820, 1024 e 1440 px | WDS-D-019: os dois lados de cada ponto de quebra, mais desktop amplo | y |
| Ordem de migração | Por módulo, depois de shell e acesso | Permite conferir módulo a módulo; detalhada em Tasks | n |

**Open questions:** none - todas resolvidas ou registradas acima; as marcadas com "n" aguardam a revisão do mantenedor.

---

## Telas

| Tela | Rota | Mudanças previstas |
| --- | --- | --- |
| Início | `/inicio` | Atalhos no padrão de DSS-06 |
| Usuários | `/admin/usuarios` | Cabeçalho de página; filtros (situação, papel) na URL; lista empilhada no celular; "Mostrar mais" pelo cursor; ações de linha em menu |
| Trocar senha | `/conta/senha` | Layout de acesso (DSS-03) |
| Plano de contas | `/financeiro/contas` | Cabeçalho de página; árvore com recuo; status com ícone; estado vazio com ação |
| Lançamentos | `/financeiro/lancamentos` | Barra de filtros na URL; prioridade de colunas (data, tipo, conta, valor líquido, status); lista empilhada no celular; ação de linha nomeada |
| Detalhe do lançamento | `/financeiro/lancamentos/[id]` | Trilha de navegação; valores com MoneyValue; status; comprovantes em lista; limite de anexo em texto natural |
| Saldo | `/financeiro/saldo` | Valor de destaque; largura estreita |
| Produtos | `/estoque/produtos` | Busca na URL; lista responsiva; estado vazio com ação |
| Detalhe do produto | `/estoque/produtos/[id]` | Trilha de navegação; saldo em destaque; movimentações em lista empilhada no celular; "Devolver" com nome que identifica a movimentação |
| Sem acesso | `/sem-acesso` | Estado padrão (DSP-02) |
| Entrar, recuperar acesso, redefinir senha | `/entrar`, `/recuperar-acesso`, `/redefinir-senha` | Layout de acesso (DSS-03) |

---

## User Stories

### DSM-01: Migração sem regressão ⭐ MVP

**User Story**: Como mantenedor, quero a interface nova sem perder nada do que a Web V1 já garante.

**Acceptance Criteria**:

1. The critérios de aceite das sub-specs 01 a 06 da Web V1 SHALL continuar verdadeiros depois da migração, exceto os explicitamente revisados em `STATE.md`.
2. The migração SHALL não alterar nenhum arquivo fora de `web/`, de `.specs/features/web-design/` e da documentação de desenvolvimento da Web.
3. The migração SHALL não acrescentar dependência a `web/package.json`.
4. IF um teste existente é alterado THEN a alteração SHALL citar a decisão WDS-D ou o requisito desta feature que a motiva, e a lista de testes alterados SHALL constar no relatório de verificação.
5. The migração SHALL não remover nem enfraquecer nenhum teste de comportamento.
6. IF o espécime não foi aprovado pelo mantenedor THEN nenhuma tela SHALL ser migrada.

**Independent Test**: rodar os gates de `web` e comparar a lista de testes antes e depois; conferir `git diff --name-only develop -- api/` vazio.

---

### DSM-02: Telas de Identity ⭐ MVP

**User Story**: Como administrador, quero gerir usuários com os mesmos padrões das outras áreas, inclusive no celular.

**Acceptance Criteria**:

1. The tela de usuários SHALL usar o cabeçalho de página, a barra de filtros e a listagem responsiva dos padrões compartilhados.
2. WHEN o filtro de situação ou de papel muda THEN a tela de usuários SHALL refletir o filtro na URL e continuar usando os filtros `active` e `role` da API.
3. The ações de cada usuário SHALL ficar num menu cujo nome acessível identifica o usuário.
4. The tela de troca de senha SHALL usar o layout das telas de acesso.
5. The telas de Identity SHALL manter os critérios de USR-01 a USR-05 e de ACS-01 a ACS-03 da Web V1.

**Independent Test**: em 390 px, filtrar por papel, abrir o menu de um usuário e executar uma ação; recarregar e ver o filtro mantido.

---

### DSM-03: Telas de Financeiro ⭐ MVP

**User Story**: Como tesouraria, quero ler valores e status sem ambiguidade e trabalhar no celular quando precisar.

**Acceptance Criteria**:

1. The telas de plano de contas, lançamentos, detalhe do lançamento, saldo e comprovantes SHALL usar o cabeçalho de página, os estados de tela e os componentes do Design System.
2. The valores monetários SHALL ser exibidos com o componente de valor monetário, em algarismos tabulares.
3. The status de lançamento `CRIADA`, `RECEBIDA`, `PAGA` e `CANCELADA` SHALL ter rótulo e ícone distintos entre si.
4. WHEN um filtro de lançamentos muda THEN a tela SHALL refletir tipo, status, conta e período na URL.
5. WHEN a largura da tela é menor que 768 px THEN a listagem de lançamentos SHALL mostrar itens empilhados com data, tipo, conta, valor líquido e status.
6. The link para o detalhe de um lançamento SHALL ter nome acessível que identifica o lançamento.
7. The plano de contas SHALL manter a hierarquia visível em qualquer largura, por recuo.
8. The telas de Financeiro SHALL manter os critérios de FWB-01 a FWB-05 da Web V1.

**Independent Test**: em 390 px, criar um lançamento, confirmá-lo, anexar um comprovante e conferir saldo; filtrar por status e compartilhar a URL.

---

### DSM-04: Telas de Estoque ⭐ MVP

**User Story**: Como responsável de estoque, quero registrar movimentações pelo celular com poucos toques e sem erro.

**Acceptance Criteria**:

1. The telas de produtos e de detalhe do produto SHALL usar o cabeçalho de página, os estados de tela e os componentes do Design System.
2. WHEN o termo de busca de produtos muda THEN a tela SHALL refletir o termo na URL.
3. WHEN a largura da tela é menor que 768 px THEN o histórico de movimentações SHALL mostrar itens empilhados com tipo, quantidade, data e responsável.
4. The ação "Devolver" SHALL ter nome acessível que identifica a movimentação de origem.
5. The diálogos de movimentação e de ajuste SHALL ocupar a tela inteira abaixo de 640 px, com as ações fixas na base.
6. The quantidades e os saldos SHALL usar algarismos tabulares, e o saldo negativo SHALL ser indicado por ícone e texto além da cor.
7. The telas de Estoque SHALL manter os critérios de EWB-01 a EWB-04 da Web V1.

**Independent Test**: em 390 px com emulação de toque, registrar uma entrada, uma saída e uma devolução a partir do histórico.

---

### DSM-05: Limpeza ⭐ MVP

**User Story**: Como quem mantém a interface, quero que não reste estilo antigo competindo com o Design System.

**Acceptance Criteria**:

1. The páginas em `web/app/` SHALL não declarar contêiner, largura nem título com classes próprias; SHALL usar o contêiner e o cabeçalho de página.
2. The código de telas SHALL não conter `<select>` fora do componente do kit.
3. The componentes de campo específicos de uma feature que duplicam o kit (`select-field` de lançamentos e o filtro de data com rótulo manual) SHALL ser substituídos pelos do kit.
4. The guarda de tokens (DSF-01 AC7) SHALL passar em todo o código de `web/app`, `web/components` e `web/features`.
5. The código SHALL não conter variantes condicionadas a tema escuro sem uso.

**Independent Test**: buscar `<select`, `max-w-` e `text-xl font-semibold` em `web/app` e `web/features` e encontrar só usos dentro do kit.

---

### DSM-06: Verificação ⭐ MVP

**User Story**: Como mantenedor, quero evidência de que a interface nova é acessível, responsiva e fiel à direção aprovada.

**Acceptance Criteria**:

1. The gates de `web` (lint, typegen, typecheck, testes, build, audit e contrato) SHALL passar.
2. The revisão com a skill `web-design-guidelines` SHALL ser executada pelo verificador final sobre `web/app`, `web/components` e `web/features`, como gate obrigatório, e cada achado SHALL ser corrigido ou registrado como conflito com decisão do projeto.
3. The conferência visual SHALL ser feita num navegador real, num build de produção, em 390, 768, 820, 1024 e 1440 px, em todas as telas da tabela "Telas", e registrada em arquivo de evidência sem senha, token nem dado pessoal.
4. The conferência SHALL incluir um percurso só por teclado de login, criação de lançamento, cancelamento com motivo e movimentação de estoque.
5. The conferência SHALL incluir ampliação a 200% e largura de 320 px, sem rolagem horizontal de página.
6. The verificação SHALL ser feita por um verificador independente do autor, com relatório em `validation.md`.
7. IF a verificação encontra divergência entre uma tela e o espécime aprovado THEN a divergência SHALL virar tarefa de correção antes do veredito.

**Independent Test**: ler `validation.md` e o arquivo de evidência; refazer um dos percursos por teclado.

---

## Edge Cases

- WHEN uma tela é aberta com a URL antiga, sem parâmetros THEN ela SHALL abrir com os filtros padrão.
- IF um registro tem texto muito longo (nome de conta, motivo, nome de produto) THEN a tela SHALL tratá-lo por truncamento ou quebra, sem quebrar o layout em 320 px.
- WHEN a pessoa com troca de senha obrigatória entra THEN ela SHALL ver só a tela de troca de senha no layout de acesso, sem navegação (FND-05 AC8).
- IF o envio de comprovante excede o limite do front THEN a mensagem SHALL usar o texto de limite definido em DSP-07.

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
| --- | --- | --- | --- |
| DSM-01 | DSM-01: Migração sem regressão | - | Pending |
| DSM-02 | DSM-02: Telas de Identity | - | Pending |
| DSM-03 | DSM-03: Telas de Financeiro | - | Pending |
| DSM-04 | DSM-04: Telas de Estoque | - | Pending |
| DSM-05 | DSM-05: Limpeza | - | Pending |
| DSM-06 | DSM-06: Verificação | - | Pending |

**Coverage:** 6 total, 0 mapped to tasks, 6 unmapped.

---

## Success Criteria

- [ ] Todas as telas da tabela "Telas" conferidas em 390, 768, 820, 1024 e 1440 px.
- [ ] `validation.md` com veredito PASS e evidência por critério.
- [ ] Nenhum achado da `web-design-guidelines` sem correção ou conflito registrado.
