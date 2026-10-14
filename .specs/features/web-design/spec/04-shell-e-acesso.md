# Shell e Acesso — Specification

## Problem Statement

A moldura da aplicação não tem identidade: cabeçalho e navegação brancos, marca em texto ("TJ Platform"), itens sem ícone, e a navegação lateral aparece já em 768 px, deixando cerca de 530 px para as tabelas. As telas de acesso (login, recuperação, redefinição e troca de senha) são um cartão centralizado sem marca. Uma falha de renderização mostra a tela de erro padrão do Next, em inglês, porque não existem `error.tsx`, `global-error.tsx` nem `not-found.tsx`. Não há link de pular para o conteúdo, nem `theme-color`. Esta sub-spec define o shell escuro, a navegação responsiva, as telas de acesso, as páginas de erro e os metadados.

## Goals

- [ ] Quem abre o sistema reconhece a Torcida Jovem pela moldura, pelo escudo e pela tipografia, antes de ler qualquer texto.
- [ ] A navegação funciona por teclado, toque e leitor de tela, do celular ao desktop.
- [ ] Nenhuma tela do sistema aparece em inglês nem sem identidade, inclusive em falha.

## Out of Scope

| Feature | Reason |
| --- | --- |
| Mudança no comportamento de login, sessão, recuperação e troca de senha | FND-03 e ACS-01 a ACS-03 da Web V1 continuam valendo. |
| Indicadores e gráficos na página inicial | WDS-D-010. |
| Navegação em trilho de ícones, busca global, notificações | Registradas como ideias adiadas ou inexistentes na V1. |
| A FERA nas telas de acesso e de erro | WDS-D-013: sem asset oficial. |
| Portal público | WDS-D-010. |

---

## Assumptions & Open Questions

| Assumption / decision | Chosen default | Rationale | Confirmed? |
| --- | --- | --- | --- |
| Moldura | Shell escuro, conteúdo claro | WDS-D-005 | y |
| Escudo no shell | Arquivo oficial, sem alteração do desenho | WDS-D-008 | y |
| Ponto de quebra da navegação fixa | 1024 px; abaixo disso, gaveta | WDS-D-019; revisa FND-05 AC2, com revalidação obrigatória | y |
| Nome ao lado do escudo | "Torcida Jovem" no shell e no login; "TJ Platform" só como título e metadado | WDS-D-020 | y |
| Favicon | O próprio escudo, só redimensionado, sem redesenho | WDS-D-021 | y |
| Ícones da navegação | Um ícone Lucide por item, escolhido no Design | Leitura rápida; biblioteca já instalada | n |
| Itens do menu da pessoa | "Trocar senha" como link; "Sair" como botão | Navegação usa link (regra da `web-design-guidelines`) | n |
| Detalhe técnico nas páginas de erro | Não exibido | Não expor estrutura interna; a pessoa só precisa do próximo passo | n |

**Open questions:** none - todas resolvidas ou registradas acima; as marcadas com "n" são suposições do autor, detalhadas em `design/04-shell-e-acesso.md`.

---

## User Stories

### DSS-01: Shell com identidade ⭐ MVP

**User Story**: Como pessoa autenticada, quero uma moldura que diga onde estou e de quem é o sistema.

**Acceptance Criteria**:

1. The shell SHALL usar os tokens do shell escuro no cabeçalho e na navegação, e os tokens de conteúdo claro na área de conteúdo.
2. The shell SHALL exibir o escudo oficial e o nome "Torcida Jovem", com link para a página inicial, e não SHALL exibir "TJ Platform" como marca.
3. The navegação SHALL agrupar os itens por módulo, com rótulo de grupo, e cada item SHALL ter ícone e texto.
4. WHEN um item corresponde à página atual ou a uma página abaixo dela THEN o item SHALL ser marcado com `aria-current`, indicador vermelho e peso de texto maior.
5. The primeiro elemento focável da página SHALL ser o link "Pular para o conteúdo", visível ao receber foco, que leva o foco à região principal.
6. The página SHALL ter uma única região principal, e as regiões de navegação SHALL ter nome acessível.
7. The menu da pessoa SHALL mostrar o nome dela, o acesso a "Trocar senha" como link e "Sair" como botão.
8. The shell SHALL manter FND-05 AC1 (itens conforme a permissão) e AC8 (só "Sair" durante a troca obrigatória de senha) da Web V1.
9. The texto do shell SHALL ter contraste de pelo menos 4,5:1 contra o fundo escuro.

**Independent Test**: abrir uma página de módulo, pressionar Tab uma vez e ativar o link de pular; conferir o item ativo com leitor de tela.

---

### DSS-02: Navegação responsiva ⭐ MVP

**User Story**: Como pessoa no celular ou no tablet, quero chegar a qualquer área sem perder espaço de trabalho.

**Acceptance Criteria**:

1. WHEN a largura da tela é de pelo menos 1024 px THEN a navegação SHALL ficar fixa ao lado do conteúdo.
2. WHEN a largura da tela é menor que 1024 px THEN a navegação SHALL ficar numa gaveta aberta por um botão com nome acessível e alvo de pelo menos 44 por 44 px.
3. WHEN a pessoa escolhe um item na gaveta THEN a gaveta SHALL fechar.
4. WHEN a gaveta fecha THEN o foco SHALL voltar ao botão que a abriu.
5. The cabeçalho SHALL permanecer visível durante a rolagem, e nenhuma ação SHALL ficar inacessível em nenhuma largura a partir de 320 px.
6. The shell SHALL respeitar as áreas seguras do dispositivo (`env(safe-area-inset-*)`).
7. WHEN a verificação deste ciclo é executada THEN o comportamento da navegação SHALL ser reconferido num navegador real em 390, 768, 820, 1024 e 1440 px, sem reaproveitar a verificação de FND-05 AC2 feita para o ponto de quebra anterior.

**Independent Test**: em 390, 768 e 820 px, abrir a gaveta, navegar e conferir o fechamento e a volta do foco; em 1024 e 1440 px, conferir a navegação fixa.

---

### DSS-03: Telas de acesso ⭐ MVP

**User Story**: Como pessoa que entra no sistema, quero reconhecer a Torcida Jovem já na tela de login.

**Acceptance Criteria**:

1. The telas de login, recuperação de acesso, redefinição de senha e troca de senha SHALL exibir o escudo oficial e o nome da torcida.
2. The telas de acesso SHALL usar o mesmo layout, com formulário em coluna única e largura estreita.
3. The telas de acesso SHALL manter todo o comportamento de FND-03 e de ACS-01 a ACS-03 da Web V1.
4. The campos de e-mail e senha SHALL manter `autocomplete` de autenticação, e o campo de e-mail SHALL desativar a correção ortográfica.
5. The telas de acesso SHALL não usar ilustração, fotografia nem A FERA.
6. WHEN a largura da tela é menor que 768 px THEN a tela de acesso SHALL mostrar o formulário sem rolagem horizontal e com os campos em largura total.

**Independent Test**: abrir `/entrar` em 390 e 1280 px; rodar os testes existentes de login e recuperação sem alteração de comportamento.

---

### DSS-04: Páginas de erro ⭐ MVP

**User Story**: Como pessoa que encontra uma falha, quero uma página em português que me diga o que fazer.

**Acceptance Criteria**:

1. IF ocorre uma exceção de renderização numa rota autenticada THEN a aplicação SHALL mostrar, dentro do shell, uma página de erro em português com as ações "Tentar de novo" e "Voltar ao início".
2. IF ocorre uma exceção de renderização no layout raiz THEN a aplicação SHALL mostrar uma página de erro em português, com o escudo e a ação "Tentar de novo".
3. WHEN a pessoa abre um endereço que não existe THEN a aplicação SHALL mostrar uma página "Página não encontrada" em português, com link para o início.
4. The páginas de erro SHALL não exibir mensagem técnica, pilha de execução nem identificador interno.
5. The páginas de erro SHALL ter um `h1` e SHALL levar o foco ao título ao aparecer.

**Independent Test**: forçar uma exceção num componente de página e abrir um endereço inexistente, num build de produção.

---

### DSS-05: Metadados e ícones

**User Story**: Como pessoa com várias abas abertas, quero identificar a aba do sistema.

**Acceptance Criteria**:

1. The documento SHALL declarar `lang="pt-BR"` e `theme-color` igual à cor do shell.
2. The título de cada página SHALL seguir o formato "Página | TJ Platform".
3. The ícone da aba SHALL ser derivado do escudo oficial, sem alteração do desenho.
4. IF o escudo redimensionado não funciona tecnicamente num contexto de ícone THEN a necessidade de um asset oficial reduzido SHALL constar em `STATE.md`, e nenhuma versão simplificada do escudo SHALL ser criada.

**Independent Test**: abrir o sistema e conferir título, ícone e cor da barra do navegador no celular.

---

### DSS-06: Página inicial

**User Story**: Como pessoa autenticada, quero chegar rápido às áreas que uso.

**Acceptance Criteria**:

1. The página inicial SHALL mostrar atalhos para as áreas que a pessoa pode ver, gerados da mesma lista da navegação (INT-01 AC4 da Web V1).
2. The atalhos SHALL ser links, com ícone, nome da área e uma frase de descrição.
3. The página inicial SHALL não mostrar indicadores, gráficos nem dados de exemplo.
4. WHEN a largura da tela é menor que 768 px THEN os atalhos SHALL ficar em uma coluna.

**Independent Test**: entrar com duas contas de permissões diferentes e conferir os atalhos de cada uma.

---

## Edge Cases

- WHEN o nome da pessoa é longo THEN o menu da pessoa SHALL truncar o nome no cabeçalho e mostrá-lo inteiro dentro do menu.
- WHEN a pessoa não tem nenhuma permissão de área THEN a navegação SHALL ficar vazia sem quebrar o layout, e a página inicial SHALL explicar que não há áreas disponíveis.
- IF a imagem do escudo não carrega THEN o nome da torcida em texto SHALL continuar visível.
- WHEN a tela é girada com a gaveta aberta e passa a ter pelo menos 1024 px THEN a gaveta SHALL fechar e a navegação fixa SHALL aparecer.

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
| --- | --- | --- | --- |
| DSS-01 | DSS-01: Shell com identidade | - | Pending |
| DSS-02 | DSS-02: Navegação responsiva | - | Pending |
| DSS-03 | DSS-03: Telas de acesso | - | Pending |
| DSS-04 | DSS-04: Páginas de erro | - | Pending |
| DSS-05 | DSS-05: Metadados e ícones | - | Pending |
| DSS-06 | DSS-06: Página inicial | - | Pending |

**Coverage:** 6 total, 0 mapped to tasks, 6 unmapped.

---

## Success Criteria

- [ ] O escudo e o nome aparecem no shell e em todas as telas de acesso.
- [ ] Os testes existentes do shell, do login e da recuperação continuam passando, alterados só pelo novo ponto de quebra (WDS-D-019) e pelo nome da marca (WDS-D-020).
- [ ] Nenhuma tela em inglês em falha de renderização ou endereço inexistente.
