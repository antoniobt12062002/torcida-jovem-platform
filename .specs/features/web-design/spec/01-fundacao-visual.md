# Fundação Visual — Specification

## Problem Statement

A Web V1 foi construída antes da identidade visual. A folha de estilos tem a paleta neutra padrão do kit, a fonte de corpo cai em Times New Roman por uma variável que referencia a si mesma (`web/app/globals.css`), não existem tokens de sucesso e aviso, e a marca é o texto "TJ Platform". Sem uma fundação única, cada tela decide cor, tamanho e largura por conta própria. Esta sub-spec define a fundação: tokens em camadas, cor, tipografia, espaço e forma, uso da marca e a página de espécime que valida tudo antes da migração das telas.

## Goals

- [ ] Toda cor, tipografia, espaço, raio e sombra da interface vem de um token; nenhuma tela usa valor solto.
- [ ] A paleta oficial (WDS-D-001) e a tipografia aprovada (WDS-D-004) estão aplicadas com contraste WCAG 2.2 AA verificado por teste.
- [ ] O mantenedor aprova a fundação no espécime antes de qualquer tela ser migrada.

## Out of Scope

| Feature | Reason |
| --- | --- |
| Tema escuro e seletor de tema | WDS-D-003. A arquitetura permite; os valores não são criados. |
| Versões novas do escudo (monocromática, negativa, reduzida) e A FERA | WDS-D-008 e WDS-D-013: dependem de asset oficial. |
| Componentes e seus estados | Sub-spec 02. |
| Migração das telas | Sub-spec 05. |
| Stitch ou outro serviço externo de prototipagem | WDS-D-009. |

---

## Assumptions & Open Questions

| Assumption / decision | Chosen default | Rationale | Confirmed? |
| --- | --- | --- | --- |
| Paleta dos tokens | `#111111`, `#D81E28`, `#FFFFFF`, `#A7A9AC` | WDS-D-001 | y |
| Papel do vermelho | Acento de identidade; ação principal em preto | WDS-D-002 | y |
| Fontes | Barlow Condensed (600, 700) e Geist, por `next/font` | WDS-D-004 | y |
| Nomes da camada semântica | Mantêm os nomes do shadcn já usados pelo kit; a família `--sidebar-*` vira a camada do shell | Trocar a paleta sem reescrever componentes | n |
| Bloco `.dark` atual | Removido | Valores nunca verificados e nenhum código o ativa; WDS-D-003 | n |
| Cores de sucesso e aviso | Verde e âmbar, restritas a estado; informação em neutro, sem azul | WDS-D-018 | y |
| Vermelho de erro | Tom distinto do vermelho de marca (proposta `#B3141D`) | Erro e marca não podem ser confundidos | n |
| Onde vive o espécime | Rota interna em `web/`, indisponível em produção | WDS-D-017 | y |
| Valores exatos de neutros, raios e alturas | Os de `context.md`, ajustáveis no Design | São proposta até o espécime ser aprovado | n |
| Algarismos tabulares na Barlow Condensed | O arquivo da fonte tem o recurso `tnum` (verificado no Design); a leitura é confirmada no espécime, e sem boa leitura os valores usam Geist | WDS-D-025 | y |

**Open questions:** none - todas resolvidas ou registradas acima; as marcadas com "n" são suposições do autor, detalhadas em `design/01-fundacao-visual.md`.

---

## User Stories

### DSF-01: Tokens em camadas ⭐ MVP

**User Story**: Como quem mantém a interface, quero os valores visuais em três camadas, para trocar a paleta ou criar um tema sem reescrever componentes.

**Acceptance Criteria**:

1. The folha de estilos global SHALL definir três camadas de tokens: primitivos (marca e escalas), semânticos (função) e de componente.
2. The tokens primitivos de marca SHALL ter exatamente os valores `#111111`, `#D81E28`, `#FFFFFF` e `#A7A9AC`.
3. The camada semântica SHALL oferecer tokens de fundo, superfície, texto, texto secundário, ação primária, acento de marca, borda decorativa, borda de campo, foco, erro, sucesso, aviso e informação, além dos tokens do shell.
4. The componentes e telas SHALL referenciar somente tokens semânticos ou de componente, nunca primitivos nem valores literais.
5. The definição de um tema SHALL ficar restrita a um único bloco que atribui valores aos tokens semânticos.
6. The aplicação SHALL entregar um único tema e não SHALL oferecer seletor de tema nem reagir a `prefers-color-scheme`.
7. IF um arquivo de tela ou de componente em `web/app`, `web/components` ou `web/features` contém cor literal (hexadecimal, `rgb`, `hsl`, `oklch`) ou utilitário de paleta crua do Tailwind THEN a guarda automatizada SHALL falhar, indicando arquivo e linha.

**Independent Test**: rodar a guarda com um arquivo de exemplo contendo `text-red-500` e outro contendo `#D81E28` e ver as duas falhas; conferir que a folha global passa.

---

### DSF-02: Cor e contraste ⭐ MVP

**User Story**: Como pessoa que usa o sistema, quero ler tudo com conforto e reconhecer a Torcida Jovem pela cor, sem a tela virar um bloco vermelho.

**Acceptance Criteria**:

1. The pares de texto e fundo formados pelos tokens semânticos SHALL ter contraste de pelo menos 4,5:1.
2. The bordas de campo, os indicadores de foco e os elementos gráficos necessários para entender um estado SHALL ter contraste de pelo menos 3:1 contra o fundo adjacente.
3. The suíte de testes SHALL calcular o contraste de cada par semântico declarado e falhar quando um par ficar abaixo do limite dos critérios 1 e 2.
4. The ação primária SHALL usar o preto de marca como fundo.
5. The vermelho de marca SHALL ser usado somente como acento: item ativo de navegação, indicador de aba ativa, filete de ênfase e marca.
6. The vermelho de marca SHALL não ser usado como fundo de página, de seção, de cartão ou de cabeçalho.
7. The token de erro SHALL ter valor diferente do vermelho de marca.
8. The aço de marca SHALL não ser usado como cor de texto nem como borda de campo sobre superfície clara.
9. The vermelho de marca SHALL não ser usado como cor de texto com menos de 24 px (ou 18,66 px em negrito) sobre o shell escuro.
10. The cores de sucesso e de aviso SHALL ser usadas somente para comunicar estado, nunca como identidade nem como cor de ação primária.

**Independent Test**: rodar o teste de contraste; abrir o espécime e contar os usos de vermelho numa tela operacional de exemplo.

---

### DSF-03: Tipografia ⭐ MVP

**User Story**: Como pessoa que usa o sistema, quero títulos com a força da torcida e texto de trabalho legível por horas.

**Acceptance Criteria**:

1. The aplicação SHALL carregar Barlow Condensed (pesos 600 e 700), Geist e Geist Mono por `next/font`, sem dependência nova em `package.json`.
2. The texto de corpo, formulários, tabelas, botões e mensagens SHALL ser renderizado em Geist.
3. The Barlow Condensed SHALL ser usada somente em título de página, título de seção, rótulo de grupo de navegação, sobretítulo, nome da marca e valor de destaque.
4. The caixa alta SHALL ser aplicada somente a título de página, rótulo de grupo e sobretítulo.
5. The interface SHALL usar somente os tamanhos da escala tipográfica declarada nos tokens, com mínimo de 12 px.
6. The colunas numéricas e os valores monetários SHALL usar algarismos tabulares.
7. WHILE o ponteiro primário é de toque THE campos de entrada SHALL usar texto de pelo menos 16 px.
8. The títulos de página e de seção SHALL usar quebra de linha equilibrada (`text-wrap: balance`).

**Independent Test**: no navegador, inspecionar a fonte computada de um parágrafo (Geist) e de um título de página (Barlow Condensed); conferir o alinhamento dos algarismos numa coluna de valores.

---

### DSF-04: Espaço, forma, elevação e movimento ⭐ MVP

**User Story**: Como pessoa que usa o sistema, quero telas com o mesmo ritmo e as mesmas proporções em todos os módulos.

**Acceptance Criteria**:

1. The interface SHALL usar somente a escala de espaçamento declarada nos tokens, de base 4 px.
2. The interface SHALL usar três raios de canto declarados nos tokens, e o formato de pílula SHALL ficar restrito a indicadores de status.
3. The superfícies em repouso SHALL não ter sombra; sombra SHALL ficar restrita a menus, diálogos, gavetas e toasts.
4. The interface SHALL não usar gradiente decorativo nem efeito de brilho.
5. The transições SHALL usar somente as durações declaradas nos tokens e SHALL listar as propriedades animadas, nunca `transition: all`.
6. WHILE o sistema operacional pede redução de movimento (`prefers-reduced-motion: reduce`) THE interface SHALL suprimir animações e transições não essenciais.
7. The áreas de conteúdo das páginas SHALL usar uma das três larguras nomeadas nos tokens (estreita, padrão ou larga).
8. The comportamento responsivo SHALL usar dois pontos de quebra, 768 px e 1024 px.

**Independent Test**: buscar `transition-all` e larguras `max-w-*` avulsas no código das telas e não encontrar; ativar redução de movimento e abrir um diálogo.

---

### DSF-05: Marca ⭐ MVP

**User Story**: Como Torcida Jovem, quero o escudo oficial presente e intacto, e nenhum elemento de marca inventado.

**Acceptance Criteria**:

1. The escudo oficial SHALL estar versionado em `web/public/brand/`, com o enquadramento do SVG ajustado ao círculo e com todos os traçados e cores do desenho idênticos aos do arquivo recebido.
2. The interface SHALL exibir o escudo sempre em proporção 1:1, sem recolorir, recortar, girar nem aplicar efeito.
3. The escudo SHALL ser exibido com pelo menos 32 px de diâmetro e com área de respiro de pelo menos um oitavo do diâmetro.
4. WHEN o escudo aparece ao lado do nome escrito THEN a imagem SHALL ter texto alternativo vazio; WHEN aparece sozinho THEN SHALL ter o texto alternativo "Torcida Jovem".
5. The imagem do escudo SHALL declarar largura e altura.
6. IF não existe asset oficial para um uso da marca (versão monocromática, negativa, reduzida ou A FERA) THEN a interface SHALL não usar substituto criado para este projeto, e a dependência SHALL constar em `STATE.md`.
7. The área operacional SHALL não exibir A FERA nem ilustração de torcida.
8. The diretório `web/public/` SHALL não conter os arquivos de exemplo do `create-next-app`.

**Independent Test**: comparar os atributos `d` e as classes de cor dos traçados do SVG versionado com os do arquivo recebido; conferir o `viewBox`.

---

### DSF-06: Página de espécime ⭐ MVP

**User Story**: Como mantenedor, quero ver a fundação e os componentes numa só página, para aprovar a direção antes da migração das telas.

**Acceptance Criteria**:

1. The página de espécime SHALL mostrar a paleta com o contraste de cada par, os tokens, a tipografia, a hierarquia de títulos, os botões, os campos, os indicadores de status, a tabela, os estados de tela, o shell com a navegação e exemplos responsivos.
2. The página de espécime SHALL ser construída somente com os tokens e os componentes reais usados pelas telas, sem componente nem estilo exclusivo do espécime.
3. The página de espécime SHALL mostrar cada componente interativo em repouso, foco, desabilitado e inválido, quando o estado existir.
4. The página de espécime SHALL mostrar a tabela de exemplo nos três comportamentos de largura (abaixo de 768 px, de 768 a 1023 px e a partir de 1024 px).
5. WHERE a aplicação roda em produção THE rota do espécime SHALL responder como página não encontrada.
6. IF o mantenedor ainda não aprovou o espécime THEN a migração das telas (sub-spec 05) SHALL não começar.

**Independent Test**: abrir o espécime em 390, 820 e 1280 px; conferir que a rota responde como não encontrada num build de produção.

---

## Edge Cases

- IF a fonte ainda não carregou THEN a interface SHALL exibir o texto com a fonte de reserva do sistema, sem texto invisível.
- WHEN o conteúdo é ampliado a 200% THEN a página SHALL continuar utilizável sem rolagem horizontal.
- IF um token semântico é referenciado por um componente e não tem valor no tema THEN a guarda automatizada SHALL falhar.

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
| --- | --- | --- | --- |
| DSF-01 | DSF-01: Tokens em camadas | - | Pending |
| DSF-02 | DSF-02: Cor e contraste | - | Pending |
| DSF-03 | DSF-03: Tipografia | - | Pending |
| DSF-04 | DSF-04: Espaço, forma, elevação e movimento | - | Pending |
| DSF-05 | DSF-05: Marca | - | Pending |
| DSF-06 | DSF-06: Página de espécime | - | Pending |

**Coverage:** 6 total, 0 mapped to tasks, 6 unmapped.

---

## Success Criteria

- [ ] O teste de contraste e a guarda de tokens passam.
- [ ] A interface renderiza em Geist e Barlow Condensed, nunca na fonte padrão do navegador.
- [ ] O espécime foi aprovado pelo mantenedor e a aprovação está registrada em `STATE.md`.
