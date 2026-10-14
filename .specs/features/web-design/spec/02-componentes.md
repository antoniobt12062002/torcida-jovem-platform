# Componentes — Specification

## Problem Statement

O kit da Web V1 (`web/components/ui/`) foi instalado com a aparência padrão e sem regra de uso. O botão padrão tem 32 px, pequeno para toque; o destrutivo é um fundo rosado, fraco para ação irreversível; `RECEBIDA` e `PAGA` têm o mesmo indicador preto; há um `<select>` nativo estilizado à mão dentro de uma feature; botão e badge usam `transition-all`; e não há componentes para cabeçalho de página, barra de filtros, lista responsiva nem valor monetário. Esta sub-spec define o conjunto de componentes, suas variantes, seus estados e seus requisitos de acessibilidade.

## Goals

- [ ] Cada componente tem variantes e estados definidos uma única vez, com foco visível, teclado e nome acessível.
- [ ] As telas montam listas, formulários e estados só com componentes do kit, sem estilo próprio.
- [ ] Nenhuma dependência nova entra em `package.json`.

## Out of Scope

| Feature | Reason |
| --- | --- |
| Componentes de gráfico e de indicador de dashboard | WDS-D-010. |
| Seletor de data próprio, editor de texto, upload com arrastar e soltar | Nenhuma tela da V1 precisa; o campo de data nativo e o envio atual continuam. |
| Storybook ou ferramenta de catálogo | Dependência nova; o espécime (DSF-06) cumpre o papel. |
| Regras de quando uma ação fica indisponível | Regra de negócio dos módulos; aqui só a forma de mostrar. |

---

## Assumptions & Open Questions

| Assumption / decision | Chosen default | Rationale | Confirmed? |
| --- | --- | --- | --- |
| Base dos componentes | `@base-ui/react` e o kit shadcn `base-nova` já instalados | WEB-D-003; sem dependência nova | y |
| Select | Nativo estilizado como componente do kit; o `Select` rico só com busca ou conteúdo rico | WDS-D-007 | y |
| Componentes sem uso na V1 (Radio, Switch, Tabs, Tooltip) | Especificados aqui; implementados só no primeiro uso real | WDS-D-022 | y |
| Alturas de controle | 32 (compacto), 36 (padrão), 44 (toque) | Densidade de desktop com alvo adequado em toque (WDS-D-014) | n |
| Botão destrutivo | Fundo sólido no tom de erro, texto branco | O estilo atual (fundo rosado) não comunica irreversibilidade | n |
| Mapeamento de status de domínio | Fica na feature dona; o componente de status do kit só conhece tons | Hoje `components/app/status-badge.tsx` importa tipos do financeiro | n |
| Duração dos toasts | Sucesso 5 s; erro e aviso 8 s; todos dispensáveis e pausados em foco ou hover | Tempo de leitura; ajustável no Design | n |

**Open questions:** none - todas resolvidas ou registradas acima; as marcadas com "n" aguardam a revisão do mantenedor.

---

## Inventário

| Componente | Camada | Situação |
| --- | --- | --- |
| Button, Input, Textarea, Checkbox, Label, Field, Badge, Alert, Dialog, AlertDialog, DropdownMenu, Sheet (gaveta), Skeleton, Table, Toast, Separator | Fundamental | Existe; ajustar a tokens, variantes e estados |
| Select (rico) | Fundamental | Existe; uso restrito (WDS-D-007) |
| Select nativo | Fundamental | Existe fora do kit (`features/financeiro/lancamentos/select-field.tsx` e `features/identity/usuarios/users-page.tsx`); promover ao kit |
| IconButton | Fundamental | Hoje é um tamanho do Button; vira componente com nome acessível obrigatório |
| Card, Breadcrumb, "Mostrar mais" | Fundamental | Criar |
| Radio, Switch, Tabs, Tooltip | Fundamental | Especificados; criados no primeiro uso real (WDS-D-022) |
| AppShell, ConfirmDialog, estados de tela, RequirePermission, NoAccess, ApiError | Composto | Existe; ajustar |
| PageHeader, PageContainer, FilterBar, lista responsiva de dados, ações de linha, layout de formulário, marca | Composto | Criar |
| StatusBadge, MoneyField, QuantityField | Domínio | Existe; ajustar |
| MoneyValue, selo de papel | Domínio | Criar |

---

## User Stories

### DSC-01: Estados, foco e alvos comuns ⭐ MVP

**User Story**: Como pessoa que usa teclado, toque ou leitor de tela, quero que todo controle responda do mesmo jeito.

**Acceptance Criteria**:

1. The componentes interativos SHALL ter aparência distinta para repouso, hover, pressionado, foco visível e desabilitado.
2. WHEN um controle recebe foco pelo teclado THEN ele SHALL exibir um indicador de foco de pelo menos 2 px, com contraste de pelo menos 3:1 contra o fundo, tanto no conteúdo claro quanto no shell escuro.
3. The indicador de foco SHALL não usar o vermelho de marca nem o tom de erro.
4. WHILE o ponteiro primário é de toque THE controles interativos SHALL ter alvo de pelo menos 44 por 44 px.
5. The controles interativos SHALL ter alvo de pelo menos 24 por 24 px em qualquer ponteiro.
6. The ícones decorativos SHALL ficar ocultos de tecnologias assistivas.
7. The componentes SHALL transmitir estado por ícone ou texto, nunca somente por cor.

**Independent Test**: percorrer o espécime só com Tab e ver o foco em todos os controles; medir os alvos numa janela de 390 px com emulação de toque.

---

### DSC-02: Ações ⭐ MVP

**User Story**: Como pessoa que usa o sistema, quero distinguir de imediato a ação principal, a secundária e a perigosa.

**Acceptance Criteria**:

1. The Button SHALL oferecer as variantes primária (fundo preto), secundária (contorno), discreta (sem fundo), destrutiva e de link, e os tamanhos compacto, padrão e de toque.
2. The variante destrutiva SHALL usar fundo sólido no tom de erro com texto de contraste de pelo menos 4,5:1.
3. The IconButton SHALL exigir um nome acessível como propriedade obrigatória.
4. WHEN a ação de um botão está em andamento THEN o botão SHALL ficar desabilitado, expor `aria-busy` e mostrar um rótulo de progresso terminado em "…".
5. The Button SHALL ser usado para ações, e navegação SHALL usar link, inclusive quando o link tem aparência de botão.

**Independent Test**: renderizar cada variante num teste e conferir papel e nome; tentar compilar um IconButton sem nome e ver o erro de tipo.

---

### DSC-03: Campos ⭐ MVP

**User Story**: Como pessoa que preenche formulários, quero campos iguais em todos os módulos, com rótulo e erro claros.

**Acceptance Criteria**:

1. The Field SHALL compor rótulo, controle, descrição e mensagem de erro, com o rótulo associado ao controle e clicável.
2. WHEN um campo tem erro THEN o Field SHALL marcar o controle com `aria-invalid`, ligar a mensagem por `aria-describedby` e mostrar a mensagem com ícone e texto.
3. The Input, a Textarea e o Select nativo SHALL ter a mesma altura, a mesma borda e os mesmos estados de foco, inválido, desabilitado e somente leitura.
4. The Select nativo SHALL ser um componente do kit, com cor de fundo e cor de texto explícitas.
5. WHERE a escolha exige busca, conteúdo rico ou interação mais complexa THE tela SHALL usar o Select rico do kit; nos demais casos SHALL usar o Select nativo.
6. The Checkbox, o Radio e o Switch SHALL formar um único alvo com o próprio rótulo.
7. The campos opcionais SHALL ser marcados com "(opcional)"; campos obrigatórios SHALL não levar marca.
8. The Switch SHALL ser usado somente para configuração de efeito imediato, e o Radio somente para escolha única entre até cinco opções visíveis.
9. WHERE um componente especificado não tem uso em nenhuma tela da Web V1 (Radio, Switch, Tabs, Tooltip) THE implementação SHALL ficar para o primeiro uso real, e o componente SHALL não aparecer no espécime.

**Independent Test**: renderizar Field com Input, Textarea e Select nativo em erro e conferir `aria-invalid`, `aria-describedby` e a altura computada igual.

---

### DSC-04: Sobreposições ⭐ MVP

**User Story**: Como pessoa que usa o sistema, quero diálogos, menus e avisos previsíveis no teclado e no celular.

**Acceptance Criteria**:

1. WHEN um diálogo, uma gaveta ou um menu abre THEN o foco SHALL ir para dentro dele, e WHEN fecha THEN SHALL voltar ao elemento que o abriu.
2. WHEN a pessoa pressiona Esc numa sobreposição sem ação em andamento THEN a sobreposição SHALL fechar.
3. The diálogos e as gavetas SHALL conter a rolagem (`overscroll-behavior: contain`).
4. WHEN a largura da tela é menor que 640 px THEN o Dialog SHALL ocupar a tela inteira, com as ações fixas na base.
5. The interface SHALL não abrir um diálogo sobre outro diálogo.
6. The Toast SHALL ser anunciado por região viva e oferecer os tons neutro, sucesso, aviso e erro, cada um com ícone.
7. The Toast SHALL poder ser dispensado e SHALL pausar a contagem enquanto tiver foco ou ponteiro sobre ele.
8. The Tooltip SHALL não ser o único meio de acesso a uma informação necessária para concluir a tarefa.
9. The AlertDialog SHALL ser usado para confirmar ações destrutivas ou sensíveis, e o Dialog para formulários e detalhes.

**Independent Test**: abrir e fechar cada sobreposição pelo teclado e conferir a volta do foco; abrir um diálogo em 390 px.

---

### DSC-05: Dados e navegação interna ⭐ MVP

**User Story**: Como pessoa que consulta registros, quero ler tabelas e status com rapidez e sem ambiguidade.

**Acceptance Criteria**:

1. The Table SHALL usar cabeçalhos de coluna semânticos, ter nome acessível e alinhar colunas numéricas à direita com algarismos tabulares.
2. The indicador de status SHALL combinar tom, ícone e rótulo, nos tons neutro, sucesso, aviso, erro e informação.
3. The status diferentes de um mesmo domínio SHALL ser distinguíveis entre si por rótulo e ícone, sem depender de cor.
4. The componente de status do kit SHALL não depender de tipos de nenhum módulo; o mapeamento de cada status para tom, ícone e rótulo SHALL ficar na feature dona.
5. The Card SHALL ser usado para agrupar conteúdo independente e SHALL não envolver tabelas, listas e formulários que já ocupam a página.
6. The Breadcrumb SHALL mostrar o caminho até a página atual, com o último item marcado como página atual e sem link.
7. The controle "Mostrar mais" SHALL informar quantos itens estão visíveis e o total.
8. The Tabs SHALL seguir o padrão de teclado de abas (setas alternam, Tab entra no painel) e marcar a aba ativa por indicador e por peso.

**Independent Test**: renderizar dois status do mesmo domínio num teste e conferir nomes e ícones distintos; ler a tabela de exemplo com leitor de tela.

---

### DSC-06: Componentes de domínio ⭐ MVP

**User Story**: Como tesouraria e estoque, quero dinheiro e quantidades exibidos e digitados sempre do mesmo jeito.

**Acceptance Criteria**:

1. The MoneyValue SHALL exibir centavos inteiros com `formatBRL`, em algarismos tabulares.
2. WHEN o valor é negativo THEN o MoneyValue SHALL exibir o sinal de menos, e a variante de saldo negativo SHALL acrescentar ícone além do tom de erro.
3. The MoneyValue SHALL oferecer as variantes neutra, de receita (verde de sucesso com sinal "+"), de despesa (neutra com sinal "−") e de saldo negativo (tom de erro com ícone), e nenhuma delas SHALL depender só de cor.
4. The MoneyField e o QuantityField SHALL manter o comportamento definido em FND-05 AC4 e AC5 da Web V1.
5. The selo de papel SHALL exibir o nome do papel em português, em tom neutro.

**Independent Test**: renderizar MoneyValue com `-1050`, `0` e `19850` e conferir texto e sinal; rodar os testes existentes de MoneyField e QuantityField sem alteração.

---

### DSC-07: Testes do kit

**User Story**: Como quem mantém a interface, quero cada componente coberto por teste de papel, nome e teclado.

**Acceptance Criteria**:

1. The componentes novos ou alterados SHALL ter teste de papel acessível, nome acessível e interação por teclado.
2. The testes de comportamento já existentes em `web/components/` SHALL continuar passando.
3. IF um teste existente precisa mudar por causa de uma decisão WDS-D THEN a alteração SHALL citar a decisão que a motiva.

**Independent Test**: rodar `pnpm test` e conferir que todo arquivo de componente em `web/components/` tem teste correspondente.

---

## Edge Cases

- WHEN o rótulo de um botão não cabe na largura disponível THEN o botão SHALL quebrar a linha ou crescer, sem cortar o texto.
- IF um IconButton aparece numa linha de lista THEN o nome acessível SHALL identificar o item da linha.
- WHEN vários toasts aparecem em sequência THEN a interface SHALL empilhá-los sem cobrir a ação principal da página no celular.
- IF o texto de uma célula é longo THEN a célula SHALL truncar com reticências ou quebrar a linha, conforme declarado pela coluna, sem alargar a tabela.

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
| --- | --- | --- | --- |
| DSC-01 | DSC-01: Estados, foco e alvos comuns | - | Pending |
| DSC-02 | DSC-02: Ações | - | Pending |
| DSC-03 | DSC-03: Campos | - | Pending |
| DSC-04 | DSC-04: Sobreposições | - | Pending |
| DSC-05 | DSC-05: Dados e navegação interna | - | Pending |
| DSC-06 | DSC-06: Componentes de domínio | - | Pending |
| DSC-07 | DSC-07: Testes do kit | - | Pending |

**Coverage:** 7 total, 0 mapped to tasks, 7 unmapped.

---

## Success Criteria

- [ ] Todos os componentes do inventário marcados como "existe" ou "criar" aparecem no espécime com seus estados.
- [ ] Nenhum `<select>` estilizado à mão fora do kit.
- [ ] `package.json` e `pnpm-lock.yaml` sem dependência nova.
