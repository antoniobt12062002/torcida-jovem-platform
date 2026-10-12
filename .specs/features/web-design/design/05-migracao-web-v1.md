# Migração da Web V1 — Design

**Spec**: `.specs/features/web-design/spec/05-migracao-web-v1.md`
**Status**: Aprovado pelo mantenedor em 2026-10-11
**Depende de**: `01` a `04` desta pasta.

Cobre: estratégia de migração, matriz de telas, impacto nos requisitos já verificados, estratégia de Verification e a proposta de unidades de execução.

---

## Estratégia de migração

### Princípios

1. **Fundação antes de tela.** Nenhuma pendência de `design-pendencias.md` é corrigida de forma avulsa (WDS-D-025). Cada uma é resolvida por um token, componente ou padrão, e chega às telas pela migração.
2. **Comportamento intocado.** Hooks, clientes da API, catálogos de erro, filtros puros, regras de permissão e conversão de dinheiro não mudam. A migração troca a casca: contêiner, cabeçalho, origem do estado dos filtros, componente de listagem e aparência.
3. **Portão humano.** O espécime é aprovado antes de qualquer tela (DSF-06 AC6).
4. **Tela inteira por vez.** Uma tela só é dada como migrada quando não resta nela contêiner próprio, `<select>` avulso, classe de paleta crua nem título montado à mão.
5. **Testes como contrato.** Testes de comportamento continuam; mudam só os listados em "Testes alterados".

### Rodadas

```
Rodada 1  DS-F   Fundação visual           tokens · fontes · escudo · guardas
Rodada 2  DS-K   Kit de componentes   ∥   DS-T  Infraestrutura compartilhada de testes
Rodada 3  DS-L   Bibliotecas auxiliares    estado na URL · consulta de mídia
Rodada 4  DS-A   Componentes compostos     página, listagem, estados, domínio, shell
Rodada 5  DS-E   Espécime visual
          ─────  PORTÃO: aprovação visual do mantenedor (obrigatória; nada começa sozinho)  ─────
Rodada 6  DS-S ∥ DS-ID ∥ DS-FA ∥ DS-FB ∥ DS-ES   shell e acesso · Identity · Financeiro A e B · Estoque
Rodada 7  DS-INT Integração                remoção de alias · guia · evidência
          VERIFY Verificação independente
```

Depois da rodada 1, a interface inteira já muda de fonte e de cor, porque os tokens semânticos mantêm os nomes do kit. Entre as rodadas 1 e 6 as telas ficam num estado intermediário (paleta nova, estrutura antiga), aceitável porque nada é publicado antes da verificação.

A rodada 6 **não começa automaticamente** quando o espécime termina: depende da aprovação visual registrada em `../STATE.md`.

### De pendência a solução

| Pendência | Resolvida por | Etapa |
| --- | --- | --- |
| D-01 fonte em Times | Token `--font-sans` | E0 |
| D-02 sem identidade | Tokens, tipografia, `BrandMark`, shell e `AccessLayout` | E0, E2, E4 |
| D-03 hierarquia e larguras | `PageContainer`, `PageHeader`, componentes de título | E2, E4 |
| D-04 tabelas no celular | `DataList` | E2, E4 |
| D-05 nomes de ação repetidos | `rowLabel` em `DataList` e `RowActions` | E2, E4 |
| D-06 estados vazios | `EmptyState` com `kind`, orientação e ação | E2, E4 |
| D-07 selects diferentes | `NativeSelect` | E1, E4 |
| D-08 código curto no título | Padrão de identificador (tipo e código em fonte monoespaçada); número legível continua lacuna de API | E4 |
| D-09 "até 9,9 MB" | Regra de texto de limite | E4 |
| D-10 erro em inglês | `error.tsx`, `not-found.tsx`, `global-error.tsx` | E4 |
| D-11 modo escuro | Decidido: não neste ciclo (WDS-D-003) | — |

---

## Matriz tela → componentes e tokens

Todas as telas autenticadas usam `ShellFrame`, `PageContainer` e `PageHeader`; a tabela lista o que é específico.

| Tela | Largura | Compostos | Fundamentais | Domínio | Tokens em destaque |
| --- | --- | --- | --- | --- | --- |
| Início `/inicio` | default | `EmptyState` | `Card`, `Link` | — | `border`, `card` |
| Usuários `/admin/usuarios` | wide | `FilterBar`, `DataList`, `RowActions`, `ShowMore`, estados | `NativeSelect`, `Button`, `Dialog`, `Checkbox`, `Textarea`, `DropdownMenu` | `StatusBadge` (ativo), `RoleBadge`, `ConfirmDialog` | `muted`, `border-strong` |
| Trocar senha `/conta/senha` | narrow ou `AccessLayout` | `FormLayout` | `Field`, `Input`, `Button`, `Alert` | — | `sidebar` (na troca obrigatória) |
| Plano de contas `/financeiro/contas` | default | estados, `RowActions` | `Button`, `Dialog`, `Field`, `Input`, `NativeSelect` | `StatusBadge` (ativo), `ConfirmDialog` | recuo por nível com a escala de espaço |
| Lançamentos `/financeiro/lancamentos` | wide | `FilterBar`, `DataList`, `ShowMore`, estados | `NativeSelect`, `Input` (data), `Button`, `Dialog` | `StatusBadge` (lançamento), `MoneyValue`, `MoneyField` | `success`, `info-subtle`, coluna `numeric` |
| Detalhe do lançamento `/financeiro/lancamentos/[id]` | default | `SectionTitle`, estados, `DataList` (comprovantes, `layout="table"` se couber) | `Breadcrumb`, `Button`, `Alert`, `Input` (arquivo) | `StatusBadge`, `MoneyValue`, `ConfirmDialog` | `destructive` só na confirmação |
| Saldo `/financeiro/saldo` | narrow | estados | `Card` | `MoneyValue` `balance` `display` | `--font-money-display`, `destructive` se negativo |
| Produtos `/estoque/produtos` | default | `FilterBar` (busca), `DataList`, `ShowMore`, estados | `Input` (busca), `Button`, `Dialog`, `Field` | `StatusBadge` (ativo) | `muted` |
| Detalhe do produto `/estoque/produtos/[id]` | default | `SectionTitle`, `DataList` (movimentações), `RowActions`, `ShowMore`, estados | `Breadcrumb`, `Button`, `Dialog`, `Textarea`, `Alert` | `QuantityField`, `StatusBadge` (saldo negativo) | `warning` no aviso de saldo negativo do ajuste |
| Sem acesso `/sem-acesso` | narrow | `NoAccess` | `Link` | — | — |
| Entrar, recuperar, redefinir | `AccessLayout` | `FormLayout` | `Field`, `Input`, `Button`, `Alert` | `BrandMark` | `sidebar`, `sidebar-foreground` |
| Erros | shell ou `AccessLayout` | — | `Button`, `Link` | `BrandMark` | — |

### Colunas por prioridade (proposta, a confirmar no espécime)

| Listagem | Prioridade 1 (sempre) | Prioridade 2 (a partir de 768) | Prioridade 3 (a partir de 1024) | Item empilhado |
| --- | --- | --- | --- | --- |
| Lançamentos | Data, conta, valor líquido, status | Tipo, forma de pagamento | Valor bruto, taxa | Título: conta. Campos: data, tipo. Valor líquido à direita. Status |
| Usuários | Nome, situação | E-mail, papéis | Acesso administrativo | Título: nome. Campos: e-mail, papéis. Status |
| Produtos | Código, nome, situação | Unidade | — | Título: nome. Campos: código, unidade. Status |
| Movimentações | Tipo, quantidade, data | Origem, responsável | Motivo, devolução de | Título: tipo e quantidade. Campos: data, responsável, motivo (três linhas) |
| Comprovantes | Arquivo, ação "Baixar" | Enviado por, data | Tamanho | Tabela pequena: continua tabela |
| Plano de contas | Árvore com recuo em qualquer largura; não usa `DataList` | | | |

Nenhuma listagem da V1 precisa de rolagem horizontal com essa distribuição. Se o espécime mostrar o contrário, a justificativa é registrada na tela (DSP-05 AC9).

---

## Impacto nos requisitos já verificados

"Preservado" quer dizer que o critério continua verdadeiro e os testes atuais continuam passando sem alteração de asserção de comportamento.

| Requisito da Web V1 | Efeito | Detalhe |
| --- | --- | --- |
| FND-01 (acesso same-origin) | Preservado | Nada em `lib/api`, `next.config.ts` ou no rewrite muda |
| FND-02 (kit de UI) | Preservado e ampliado | Mesmos componentes, mais os novos; AC2 (erro ligado ao campo) reforçado por teste |
| FND-03 (sessão, login, logout) | Preservado | Só layout e atributos de campo |
| FND-04 (proteção e permissões) | Preservado | `RequirePermission` inalterado |
| FND-05 AC1, AC3 a AC8 | Preservados | Shell, estados, campos e confirmação mantêm o comportamento |
| **FND-05 AC2** (navegação recolhível) | **Revisado e reverificado** | Ponto de quebra passa de 768 para 1024 px (WDS-D-019). Teste do shell e smoke refeitos em 390, 768, 820, 1024 e 1440 |
| FND-06 (erros da API) | Preservado | Mensagens e catálogos iguais; muda a apresentação (ícone) |
| ACS-01 a ACS-03 | Preservados | Layout de acesso |
| USR-01 a USR-05 | Preservados | USR-01: filtros passam a refletir na URL, continuando a usar `active` e `role` da API |
| FWB-01, FWB-02, FWB-04 | Preservados | Apresentação |
| FWB-03 (lançamentos) | Preservado | AC1: filtros no cliente continuam; estado passa à URL |
| FWB-05 (comprovantes) | Preservado, com texto revisado | AC3: o limite do front não muda (10.420.224 bytes); o texto exibido passa de "9,9 MB" a "9 MB" (DSP-07 AC6) |
| EWB-01 a EWB-04 | Preservados | EWB-02 AC3: "Devolver" continua no histórico, com nome acessível que identifica a movimentação |
| INT-01 (navegação e início) | Preservado | Mesma lista, mesma ordem, mesmas permissões; entram ícones e descrições |
| INT-02 (guia) | Atualizado | `docs/development/web.md` ganha a seção do Design System e do espécime |
| INT-03 (smoke) e INT-04 (verificação) | Refeitos para este ciclo | Evidência e relatório próprios em `.specs/features/web-design/` |
| WEB-D-005 | Revisada | Nota já registrada em `web/STATE.md` |

### Testes alterados (previsão)

Toda alteração cita a decisão ou o requisito e entra no relatório de verificação (DSM-01 AC4).

| Teste | Alteração | Motivo |
| --- | --- | --- |
| `components/app/app-shell.test.tsx` | Marca "Torcida Jovem"; link de pular; menu "Trocar senha" como link | WDS-D-020, DSS-01 AC5 e AC7 |
| `components/app/status-badge.test.tsx` | Recebe definição de status | DSC-05 AC4 |
| `components/app/confirm-dialog.test.tsx` | `confirmLabel` obrigatório | DSP-03 AC3 |
| `components/ui/display.test.tsx` | Nomes de variante | `02-componentes.md` |
| `features/financeiro/lancamentos/lancamentos-page.test.tsx` | Link "Detalhes" com nome que identifica o lançamento; filtros lidos da URL | DSP-05 AC2 e AC10 |
| `features/estoque/movimentacoes/movimentacoes-panel.test.tsx` | Botão "Devolver" com nome que identifica a movimentação | DSP-05 AC10 |
| `features/financeiro/comprovantes/comprovantes-panel.test.tsx` | Texto do limite | DSP-07 AC6 |
| `features/identity/usuarios/*` e `features/estoque/produtos/*` (testes de filtro e busca) | Estado lido da URL | DSP-05 AC2 |
| `lib/session/navigation-mock.ts` | Mock reativo ao histórico | `03-padroes.md` |

Nenhum teste é removido. O número de testes só cresce.

---

## Estratégia de Verification

### Gates por tarefa (iguais aos da Web V1)

| Gate | Comando |
| --- | --- |
| Web | `pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build` |
| Contrato | `pnpm lint:api && pnpm gen:api:check` |
| Backend intacto | `git diff --name-only develop -- api/` vazio |
| Dependências intactas | `git diff --name-only develop -- web/package.json web/pnpm-lock.yaml` vazio |
| Propriedade | Diff só nos caminhos da unidade |
| Audit | `pnpm audit --audit-level high`, na última tarefa de cada unidade |

As guardas novas (tokens, contraste, escudo) rodam dentro de `pnpm test`, então viram gate automático a partir de E0.

### Verificador final (agente independente, autor diferente do verificador)

| Passo | O que faz | Evidência |
| --- | --- | --- |
| 1. Checagem ancorada na spec | Para cada AC de DSF, DSC, DSP, DSS e DSM, localiza o teste ou a evidência e confirma que o valor afirmado é o da spec | `validation.md`, com `arquivo:linha` |
| 2. Regressão da Web V1 | Confere que os testes das seis sub-specs da Web V1 passam e que só os testes da lista acima mudaram | Lista de testes antes e depois |
| 3. Reverificação de FND-05 AC2 | Navegação em 390, 768, 820, 1024 e 1440 px, num navegador real | Capturas e registro |
| 4. **Gate `web-design-guidelines`** | Executa a skill sobre `web/app`, `web/components` e `web/features`. Cada achado é corrigido ou casado com um conflito registrado em `../STATE.md`. Achado sem correção nem conflito reprova | Saída da skill em `arquivo:linha`, anexada |
| 5. Sensor de discriminação | Injeta falhas mínimas num worktree temporário e confirma que algum teste cai. Falhas previstas abaixo | Resultado por mutante |
| 6. Conferência visual | Build de produção, stack real, todas as telas da matriz nas cinco larguras; comparação com o espécime aprovado | Arquivo de evidência, sem senha, token ou dado pessoal |
| 7. Teclado | Login, criação de lançamento, cancelamento com motivo e movimentação de estoque, só por teclado | Registro do percurso |
| 8. Zoom e refluxo | 200% e 320 px, sem rolagem horizontal de página | Registro |
| 9. Movimento | Com `prefers-reduced-motion`, abrir diálogo, gaveta e toast | Registro |

**Falhas mínimas do sensor** (cada uma deve derrubar ao menos um teste):

1. Trocar um token semântico por um valor que quebra o contraste.
2. Usar uma cor literal numa tela.
3. Alterar um traçado do escudo.
4. Remover o nome acessível de um `IconButton`.
5. Remover o `rowLabel` de uma ação de linha.
6. Tirar `aria-invalid` ou `aria-describedby` do `Field`.
7. Fazer `useQueryState` aceitar um valor inválido da URL.
8. Deixar de limpar o parâmetro ao voltar ao valor padrão.
9. Voltar o ponto de quebra do shell para `md`.
10. Remover o link de pular.
11. Fazer `MoneyValue` omitir o sinal ou o ícone do saldo negativo.
12. Tornar o espécime acessível em produção.
13. Mostrar "Mostrar mais" com exatamente 50 registros.
14. Dar a dois status do mesmo domínio o mesmo rótulo e ícone.

O ciclo de correção e nova verificação é limitado a três iterações antes de escalar ao mantenedor, como na Web V1. O fechamento exige `validate_state.py` com saída zero.

### O que os testes automatizados não cobrem

jsdom não aplica CSS. Fonte renderizada, contraste percebido, pontos de quebra, alvo de toque e rolagem horizontal só são confirmados no navegador (passos 3, 6, 8 e 9). Foi assim que o smoke visual da Web V1 achou o defeito que quatro verificadores não viram; por isso a conferência no navegador é obrigatória, e não opcional.

---

## Unidades de execução e propriedade de arquivos (DD-6, aprovada com a revisão pós-Graphify, WDS-D-028)

Caminhos relativos a `web/`. Esta tabela substitui a proposta original. Foi revisada com o diagnóstico estrutural do Graphify (0.9.86 local, só arestas `EXTRACTED`) e cada relação relevante foi conferida com os `import` reais.

Regras gerais, herdadas da Web V1: duas unidades só rodam juntas se não compartilham arquivo; cada agente trabalha em worktree e branch próprios; o coordenador integra, confere a propriedade e roda os gates antes da rodada seguinte; `package.json` e `pnpm-lock.yaml` não pertencem a nenhuma unidade neste ciclo.

| Unidade | Rodada | Arquivos exclusivos | Depende de | Paralela com |
| --- | --- | --- | --- | --- |
| **DS-F** Fundação visual | 1 | `app/globals.css`, `app/layout.tsx`, `app/icon.svg`, `app/apple-icon.png`, `app/favicon.ico` (remoção), `public/`, `test/design/`, `lib/design/` | — | — |
| **DS-K** Kit de componentes | 2 | `components/ui/` | DS-F | DS-T |
| **DS-T** Infraestrutura compartilhada de testes | 2 | `test/msw/fixtures.ts`, `test/msw/server.ts`, `test/render.tsx`, `test/setup.ts`, demais arquivos de `test/` fora de `test/design/`, `lib/session/navigation-mock.ts`, `lib/session/test-utils.tsx` | DS-F | DS-K |
| **DS-L** Bibliotecas auxiliares | 3 | `lib/url-state/`, `lib/responsive/` | DS-T | — |
| **DS-A** Componentes compostos | 4 | `components/app/` (exceto `nav-items.ts`) | DS-K, DS-L, DS-T | — |
| **DS-E** Espécime visual | 5 | `app/(dev)/`, `.specs/features/web-design/evidence/especime.md` | DS-A | — |
| *Portão* | — | Aprovação visual do mantenedor, registrada em `../STATE.md` | DS-E | — |
| **DS-S** Shell e acesso | 6 | `app/(app)/layout.tsx`, `app/(app)/error.tsx`, `app/(app)/inicio/`, `app/(app)/sem-acesso/`, `app/(app)/conta/`, `app/(public)/`, `app/not-found.tsx`, `app/global-error.tsx`, `features/identity/acesso/`, `components/app/nav-items.ts`, `.specs/features/web-design/evidence/shell-responsivo.md` | Portão | DS-ID, DS-FA, DS-FB, DS-ES |
| **DS-ID** Identity | 6 | `features/identity/usuarios/`, `features/identity/nav.ts`, `app/(app)/admin/` | Portão | as outras quatro |
| **DS-FA** Financeiro A | 6 | `features/financeiro/contas/`, `features/financeiro/saldo/`, `features/financeiro/nav.ts`, `app/(app)/financeiro/contas/`, `app/(app)/financeiro/saldo/` | Portão | as outras quatro |
| **DS-FB** Financeiro B | 6 | `features/financeiro/lancamentos/`, `features/financeiro/comprovantes/`, `app/(app)/financeiro/lancamentos/` | Portão | as outras quatro |
| **DS-ES** Estoque | 6 | `features/estoque/`, `app/(app)/estoque/` | Portão | as outras quatro |
| **DS-INT** Integração | 7 | Remoção de alias em `components/ui/` e `components/app/`; guarda global em `test/design/`; `docs/development/web.md`; `.specs/features/web-design/evidence/` (exceto os arquivos do espécime e do shell) | as cinco frentes | — |
| **VERIFY** | 7 | `.specs/features/web-design/validation.md` | DS-INT | — |

**Sem dono e sem alteração neste ciclo**: `lib/api/`, `lib/forms/`, `lib/money.ts`, `lib/utils.ts`, `lib/session/` (exceto os dois arquivos de DS-T), `app/page.tsx`, `next.config.ts`, `vitest.config.ts`, `eslint.config.mjs`, `package.json`, `pnpm-lock.yaml`. Uma tarefa que precise tocar um deles para e reporta.

### Componentes centrais e seus donos

Arquivos importados por três ou mais frentes paralelas (contagem de arquivos importadores, conferida com os `import`):

| Arquivo | Importado por | Dono | Compatibilidade temporária |
| --- | --- | --- | --- |
| `components/ui/button.tsx` | 29 arquivos, cinco frentes | DS-K | Variantes `default`, `outline` e tamanho `icon-sm` mantidos como alias |
| `components/ui/alert.tsx` | 17, cinco frentes | DS-K | `variant="destructive"` mantido como alias do tom de erro |
| `components/ui/toast.tsx` | 17, quatro frentes e `lib/session` | DS-K | API imperativa inalterada; `toast.warning` só acrescenta |
| `components/ui/input.tsx` | 12, cinco frentes | DS-K | Propriedades inalteradas |
| `components/ui/table.tsx` | 5, quatro frentes | DS-K | Rolagem embutida mantida como modo legado |
| `components/app/states.tsx` | 10, cinco frentes | DS-A | Propriedades atuais mantidas; `kind` e `shape` são opcionais |
| `components/app/status-badge.tsx` | 4, três frentes | DS-A | Aceita a forma antiga (`status` como texto) até a integração |
| `components/app/confirm-dialog.tsx` | 4, três frentes | DS-A | Propriedades atuais mantidas; as novas são opcionais |
| `test/msw/fixtures.ts`, `test/msw/server.ts` | 41 e 21, cinco frentes | DS-T | Só acréscimos; nenhuma fábrica existente muda de assinatura |
| `lib/session/navigation-mock.ts`, `lib/session/test-utils.tsx` | 33 e 36, cinco frentes | DS-T | API atual mantida; o mock passa a reagir ao histórico |

### Congelamento e exceções

Depois do portão do espécime, os arquivos de **DS-T, DS-K, DS-A e DS-L ficam congelados** para as cinco frentes da rodada 6. Cada agente consome esses arquivos e não os modifica.

Se uma frente precisar de mudança num arquivo congelado:

1. interrompe a tarefa dependente;
2. reporta a necessidade ao coordenador, com o arquivo e a mudança pedida;
3. o coordenador cria uma unidade de integração com responsável único;
4. a mudança é implementada e verificada (gates completos), de forma compatível com as outras frentes;
5. o coordenador libera a continuação das frentes afetadas, que partem do novo ponto de integração.

Os alias de compatibilidade ficam até a integração final. A remoção é de DS-INT, a única unidade autorizada a tocar arquivos de DS-K, DS-A e DS-T depois do portão.

Fixtures e utilitários de teste específicos de um módulo vão para o `test-helpers.ts` do próprio módulo, que já existe em cada um.

### O que roda em paralelo, e por quê

| Rodada | Unidades | Fundamento |
| --- | --- | --- |
| 2 | DS-K ∥ DS-T | Arquivos disjuntos. Nenhum teste de `components/ui` importa fixtures, servidor MSW, mock de navegação ou `test-utils`. `lib/session` importa `toast`, cuja mudança é só aditiva |
| 6 | DS-S ∥ DS-ID ∥ DS-FA ∥ DS-FB ∥ DS-ES | Nenhum `import` real entre as cinco. As duas arestas do grafo entre elas eram `INFERRED` (coincidência de nome entre testes) e foram descartadas. Tudo o que compartilham está congelado |

Único ponto de contato estrutural na rodada 6: `components/app/nav-items.ts` (DS-S) importa `features/*/nav.ts` dos módulos. Os campos novos de `NavItem` (`icon`, `description`) são opcionais, então a ordem de integração entre DS-S e os módulos é livre.

As demais rodadas são sequenciais por dependência real: DS-L precisa do mock reativo de DS-T; DS-A precisa de kit, lib e infraestrutura de testes; DS-E precisa de DS-A.

### Uso do Graphify neste ciclo

Ferramenta auxiliar, conforme `docs/development/ai-environment.md`. Specs e `STATE.md` têm precedência; o grafo não vira requisito nem prova de correção.

- Executável local 0.9.86, resolvido pelo caminho e com a versão conferida; nunca o do `PATH` nem o servidor MCP global.
- Atualização manual pelo coordenador: antes de fechar a DAG (feita em 2026-10-11), depois de integrar unidades que alterem dependências (rodadas 2, 4 e 6) e antes da revisão estrutural final (rodada 7).
- Agentes em paralelo não atualizam o grafo.
- Só arestas `EXTRACTED` confirmadas com o `import` real contam. Arquivos `.d.ts` não estão no grafo; ausência de aresta não prova independência.

### Gates por rodada

| Rodada | Gates |
| --- | --- |
| Toda tarefa | Web (`pnpm lint`, `next typegen`, `tsc --noEmit`, `pnpm test`, `pnpm build`), Contrato, Backend intacto, Dependências intactas, Propriedade |
| Início de cada rodada | `validate_spec.py` nas cinco specs e `validate_tasks.py` nos cinco arquivos de tasks, ambos com saída zero |
| 1 | Guardas de fundação passando (tokens da folha global, contraste, integridade do escudo) |
| 2 | Testes existentes de `components/` passando; os 33 arquivos que usam o mock de navegação passando sem alteração; guarda de tokens em `components/ui` |
| 3 | Teste de renderização no servidor do hook de consulta de mídia (SSR e hidratação) |
| 4 | Revisão `web-design-guidelines` sobre `components/` pelo coordenador; guarda de tokens em `components/app` |
| 5 | Espécime conferido em 390, 768, 820, 1024 e 1440 px; rota indisponível em produção; **aprovação do mantenedor** |
| 6 | Por frente: testes de comportamento existentes passando, só com as alterações listadas; guarda de tokens nos diretórios da frente; conferência no navegador das telas da frente |
| 7 | Grafo atualizado e revisão estrutural; remoção de alias com guarda global; smoke completo no navegador; revalidação de FND-05 AC2; `web-design-guidelines`; verificador independente; `validate_state.py` |

Em toda rodada confirma-se que regras de negócio, permissões e contratos da API permanecem intactos: gates Contrato e Backend intacto, e nenhuma alteração em `lib/api/`, `lib/forms/` e nos hooks de dados das features além da origem do estado dos filtros.

## Risks & Concerns

| Concern | Location | Impact | Mitigation |
| --- | --- | --- | --- |
| Estado intermediário entre E0 e E4 | Todas as telas | Aparência mista durante a execução | Nada é publicado antes de E6; o espécime é a referência nesse período |
| Unidades de módulo divergirem na aplicação dos padrões | E4 | Telas parecidas, mas não iguais | Os padrões são componentes com API fechada em E2; o verificador compara cada tela com o espécime |
| Alias temporários esquecidos | `components/ui/button.tsx`, `components/app/status-badge.tsx` | Duas formas de fazer a mesma coisa | DS-INT remove; a guarda de tokens recusa as variantes antigas ao final |
| Testes de comportamento enfraquecidos durante a migração | `features/**/*.test.tsx` | Regressão não detectada | Lista fechada de testes alterados; o verificador compara antes e depois |
| Smoke depende de stack local com S3 e e-mail `log` | Ambiente | Parte do percurso não executável | Mesmas ressalvas e o mesmo procedimento de `docs/development/web.md`; redefinição com token válido continua coberta por MSW |
| Conferência visual manual é demorada (12 telas em 5 larguras) | E6 | Verificação superficial por cansaço | Roteiro por tela no arquivo de evidência; capturas por script temporário fora do repositório, como no smoke visual da Web V1 |
| Arquivo do escudo, documentação e `.gitignore` ainda sem commit | Árvore de trabalho da branch `feature/web-design` | Worktrees das unidades partiriam sem a base | Antes da rodada 1, a base documental (specs, design, tasks, escudo, `ai-environment.md`, `.gitignore`) entra em commits próprios, mediante autorização do mantenedor |

## Tech Decisions

| Decisão | Escolha | Motivo |
| --- | --- | --- |
| Ordem | Fundação, kit, compostos, espécime, telas | Cada etapa só depende da anterior; o portão fica antes do trabalho mais caro |
| Paralelismo | Cinco frentes depois do portão | Mesma divisão por diretório que funcionou na Web V1 |
| Mudanças de API de componente | Alias temporários removidos na integração | Preserva a propriedade de arquivos entre unidades paralelas |
| Gate de acessibilidade | `web-design-guidelines` no verificador final | Decisão do mantenedor (WDS-D-025) |
| Infraestrutura de testes | Unidade própria (DS-T), antes das frentes paralelas | É a parte mais central do front (41 e 33 arquivos dependentes) e não pode ser alterada por várias frentes ao mesmo tempo |
| Conferência visual | Navegador real, build de produção | jsdom não aplica CSS |
