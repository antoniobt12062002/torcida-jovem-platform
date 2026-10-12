# Padrões Compartilhados — Specification

## Problem Statement

Cada tela da Web V1 resolve por conta própria o que deveria ser padrão. Há seis larguras de página diferentes, o título `h1` é repetido à mão em 16 lugares, os estados vazios são uma frase sem ação, filtros e busca vivem só em memória e se perdem na recarga, tabelas só rolam na horizontal no celular, e ações de linha repetem o mesmo nome acessível ("Detalhes", "Devolver"). Esta sub-spec define os padrões que Identity, Financeiro, Estoque e os módulos futuros compartilham: estrutura de página, estados de tela, feedback e confirmação, formulários, listagens, permissões e texto.

## Goals

- [ ] Uma pessoa que aprendeu uma tela reconhece as outras: mesma estrutura, mesmos estados, mesmas confirmações.
- [ ] Listagens funcionam em 390 px sem depender de rolagem horizontal, e o estado delas sobrevive a recarga e a compartilhamento de link.
- [ ] Toda situação de erro, vazio e falta de acesso diz à pessoa o que fazer em seguida.

## Out of Scope

| Feature | Reason |
| --- | --- |
| Paginação e filtro no servidor | WEB-D-016 e WDS-D-012: a API não muda. |
| Virtualização de listas | WDS-D-012. |
| Aviso de alterações não salvas | Registrado em `context.md` como ideia adiada. |
| Quais ações ficam indisponíveis em registros inativos ou cancelados | Regra de negócio dos módulos. |
| Novos erros ou novas mensagens da API | FND-06 da Web V1 continua valendo sem mudança. |

---

## Assumptions & Open Questions

| Assumption / decision | Chosen default | Rationale | Confirmed? |
| --- | --- | --- | --- |
| Tabelas em telas estreitas | Lista empilhada abaixo de 768 px; colunas por prioridade de 768 a 1023 px; tabela completa a partir de 1024 px | WDS-D-006 | y |
| Estado na URL | Filtros, busca, ordenação e aba na query string, com APIs nativas do Next | WDS-D-011 | y |
| Listas longas | "Mostrar mais" no cliente | WDS-D-012 | y |
| Tamanho do lote de "Mostrar mais" | 50 itens | A `web-design-guidelines` indica 50 como limite para lista sem tratamento | n |
| Quantidade exibida por "Mostrar mais" na URL | Não entra na URL | Recarregar volta ao primeiro lote; evita links que abrem listas enormes | n |
| Histórico do navegador | Filtros e busca substituem a entrada atual; troca de aba cria entrada nova | Voltar não deve desfazer letra por letra de uma busca | n |
| Tabela pequena no celular | Até 4 colunas curtas, cabendo em 320 px, continua tabela | WDS-D-006: não prejudicar tabelas pequenas | n |
| Identificador de registro sem número legível | Tipo seguido do código curto em fonte monoespaçada | WEB-D-014: o número legível exige mudança na API | n |
| Limite de anexo exibido | Inteiro arredondado para baixo ("até 9 MB") | "9,9 MB" é pouco natural; arredondar para baixo nunca promete mais que o limite real | n |
| Validação | Ao sair do campo e ao enviar | Padrão já usado com React Hook Form; a API continua a autoridade | n |

**Open questions:** none - todas resolvidas ou registradas acima; as marcadas com "n" aguardam a revisão do mantenedor.

---

## User Stories

### DSP-01: Estrutura de página ⭐ MVP

**User Story**: Como pessoa que navega entre módulos, quero que toda página comece do mesmo jeito.

**Acceptance Criteria**:

1. The páginas autenticadas SHALL usar um contêiner de página com uma das três larguras nomeadas (estreita, padrão ou larga).
2. The páginas SHALL ter um cabeçalho de página com exatamente um `h1`, uma descrição opcional e a área de ações.
3. The cabeçalho de página SHALL ter no máximo uma ação primária; as demais ações SHALL ser secundárias ou ficar em menu.
4. WHERE a página é de detalhe de um registro THE cabeçalho SHALL mostrar a trilha de navegação até a listagem de origem.
5. The níveis de título de uma página SHALL seguir a ordem `h1`, `h2`, `h3`, sem salto.
6. WHEN a largura da tela é menor que 768 px THEN as ações do cabeçalho SHALL ficar abaixo do título, e a ação primária SHALL ocupar a largura disponível.

**Independent Test**: abrir duas páginas de módulos diferentes e conferir a mesma estrutura de cabeçalho, um único `h1` e a largura nomeada.

---

### DSP-02: Estados de tela ⭐ MVP

**User Story**: Como pessoa que usa o sistema, quero saber se a tela está carregando, vazia, com erro ou fora do meu acesso, e o que fazer.

**Acceptance Criteria**:

1. WHILE os dados de uma área estão carregando THE área SHALL mostrar um esqueleto com o formato aproximado do conteúdo e anunciar o carregamento a tecnologias assistivas.
2. WHEN uma listagem não tem nenhum registro THEN a tela SHALL mostrar um estado vazio com título, uma frase de orientação e, para quem tem a permissão de criar, a ação de criar o primeiro registro.
3. WHEN os filtros ou a busca não retornam resultado THEN a tela SHALL mostrar um estado vazio diferente do anterior, com a ação "Limpar filtros".
4. IF o carregamento falha THEN a área SHALL mostrar a mensagem de erro com ícone e a ação "Tentar de novo".
5. IF a pessoa não tem a permissão de leitura da área THEN a página SHALL mostrar o estado "Sem acesso" com caminho de volta ao início.
6. The estados de tela SHALL ocupar o lugar do conteúdo, sem deslocar o cabeçalho de página.

**Independent Test**: com MSW, forçar lista vazia, filtro sem resultado, erro de carregamento e falta de permissão numa mesma listagem.

---

### DSP-03: Feedback e confirmação ⭐ MVP

**User Story**: Como pessoa que executa uma ação, quero saber se deu certo e ser avisada antes de algo irreversível.

**Acceptance Criteria**:

1. WHEN uma ação de escrita termina com sucesso THEN a interface SHALL mostrar um toast de sucesso e refletir a mudança na tela sem recarga.
2. WHEN uma ação destrutiva ou sensível é acionada THEN a interface SHALL abrir uma confirmação que nomeia o registro afetado e descreve a consequência.
3. The botão de confirmar SHALL repetir o verbo da ação (por exemplo "Cancelar lançamento"), nunca um rótulo genérico como "Confirmar" ou "OK".
4. The confirmação com motivo obrigatório SHALL manter o comportamento de FND-05 AC6 da Web V1.
5. IF a ação confirmada falha THEN a confirmação SHALL continuar aberta e mostrar a mensagem de erro dentro dela.
6. IF uma ação está indisponível por causa do estado do registro THEN o controle SHALL aparecer desabilitado, com o motivo em texto visível junto dele.
7. The mensagens de erro da API SHALL seguir FND-06 da Web V1, sem mudança de conteúdo.

**Independent Test**: cancelar um lançamento e conferir o nome do registro, a consequência e o rótulo do botão; forçar erro na confirmação com MSW.

---

### DSP-04: Formulários ⭐ MVP

**User Story**: Como pessoa que cadastra dados, no computador ou no celular, quero formulários curtos, claros e que apontem o erro.

**Acceptance Criteria**:

1. The formulários SHALL usar coluna única, com o rótulo acima do campo.
2. WHEN a pessoa sai de um campo com valor inválido THEN o formulário SHALL mostrar o erro junto do campo.
3. WHEN a pessoa envia o formulário com erros THEN o foco SHALL ir para o primeiro campo inválido.
4. IF a API devolve um erro que não pertence a nenhum campo THEN o formulário SHALL mostrá-lo num alerta no topo do formulário.
5. WHILE o envio está em andamento THE formulário SHALL desabilitar a ação de enviar e mostrar o rótulo de progresso.
6. The campos SHALL declarar `autocomplete` e tipo ou modo de entrada adequados ao dado.
7. The campos de e-mail e de código SHALL desativar a correção ortográfica.
8. The formulários SHALL não bloquear a ação de colar.
9. WHEN a largura da tela é menor que 768 px THEN as ações do formulário SHALL ocupar a largura disponível, com a ação principal por último na ordem de leitura.

**Independent Test**: enviar um formulário vazio e conferir o foco no primeiro erro; devolver `422` com MSW e conferir os campos marcados.

---

### DSP-05: Listagens ⭐ MVP

**User Story**: Como pessoa que consulta registros, quero filtrar, buscar e ler a lista em qualquer tela, e poder mandar o link da consulta para outra pessoa.

**Acceptance Criteria**:

1. The listagens com filtro ou busca SHALL ter uma barra de filtros acima da lista, com a contagem de resultados e a ação "Limpar filtros" visível quando houver filtro ativo.
2. WHEN a pessoa altera filtro, busca, ordenação ou aba THEN a interface SHALL atualizar a query string da URL.
3. WHEN a página é aberta com parâmetros na query string THEN a listagem SHALL restaurar filtro, busca, ordenação e aba a partir deles.
4. IF um parâmetro da query string tem valor inválido THEN a listagem SHALL ignorá-lo e usar o valor padrão, sem erro.
5. WHEN a largura da tela é de pelo menos 1024 px e a área de conteúdo tem pelo menos 896 px THEN a listagem SHALL mostrar a tabela completa; com a área de conteúdo menor, as colunas de prioridade 3 SHALL ser ocultadas, e o documento SHALL não rolar na horizontal.
6. WHEN a largura da tela está entre 768 e 1023 px THEN a listagem SHALL mostrar as colunas de maior prioridade declarada, e os dados ocultos SHALL continuar acessíveis no detalhe do registro.
7. WHEN a largura da tela é menor que 768 px THEN a listagem SHALL mostrar cada registro como item empilhado, com título, até três campos-chave, status e ações.
8. WHERE a tabela tem até quatro colunas curtas e cabe em 320 px THE listagem SHALL manter a tabela também abaixo de 768 px.
9. IF uma tabela precisa de rolagem horizontal THEN a spec da tela SHALL registrar a justificativa, a primeira coluna SHALL ficar fixa e a região rolável SHALL ter nome acessível e ser alcançável pelo teclado.
10. The ações de linha SHALL ter nome acessível que identifica o registro da linha.
11. The listagens SHALL mostrar uma ação de linha visível e reunir as demais num menu.
12. WHEN uma listagem carregada inteira no cliente tem mais de 50 registros THEN a interface SHALL mostrar os 50 primeiros e o controle "Mostrar mais", que acrescenta 50 por vez.
13. The filtro e a busca SHALL considerar todos os registros carregados, não só os visíveis.
14. WHERE a API oferece cursor de paginação THE controle "Mostrar mais" SHALL buscar a página seguinte pelo cursor.

**Independent Test**: filtrar lançamentos, copiar a URL, abrir em outra aba e ver o mesmo resultado; abrir a mesma listagem em 390, 768, 820, 1024 e 1440 px.

---

### DSP-06: Permissões na interface ⭐ MVP

**User Story**: Como pessoa com acesso limitado, quero ver só o que posso fazer.

**Acceptance Criteria**:

1. The exibição de ações por permissão SHALL manter FND-04 da Web V1: ação sem permissão não aparece.
2. WHILE a pessoa tem só permissão de leitura numa área THE área SHALL não mostrar controles de escrita, nem desabilitados.
3. IF a API nega uma ação que a interface exibiu THEN a interface SHALL mostrar a mensagem de permissão e manter o estado exibido.
4. The estado "Sem acesso" SHALL ser o mesmo em todas as áreas.

**Independent Test**: entrar com uma conta de `CONSELHO_FISCAL` e conferir que nenhuma tela mostra ação de escrita.

---

### DSP-07: Texto da interface

**User Story**: Como pessoa que usa o sistema, quero textos diretos, em português, que digam o que fazer.

**Acceptance Criteria**:

1. The textos da interface SHALL estar em português do Brasil, em frase comum, sem Title Case.
2. The rótulos de ação SHALL começar por verbo e dizer o objeto (por exemplo "Novo lançamento", "Desativar conta").
3. The mensagens de erro escritas pelo front SHALL dizer o que aconteceu e o próximo passo.
4. The textos SHALL usar o caractere de reticências "…" e espaço não separável entre número e unidade.
5. The rótulos de progresso SHALL terminar em "…".
6. The limite de tamanho de anexo SHALL ser exibido em megabytes inteiros, arredondado para baixo.
7. The identificador de um registro sem número legível SHALL ser exibido como o tipo seguido do código curto em fonte monoespaçada.

**Independent Test**: percorrer as telas e conferir rótulos de botões, mensagens de erro e o texto do limite de anexo.

---

## Edge Cases

- WHEN a pessoa usa voltar no navegador depois de trocar de aba THEN a interface SHALL voltar à aba anterior.
- WHEN a busca muda a cada tecla THEN a interface SHALL atualizar a URL sem criar uma entrada de histórico por tecla.
- IF um filtro restaurado da URL aponta para um registro que não existe mais (por exemplo uma conta) THEN a listagem SHALL ignorar o filtro.
- WHEN um item empilhado tem texto longo (por exemplo um motivo) THEN o item SHALL limitar o texto a três linhas, com o texto completo no detalhe.
- IF uma listagem tem exatamente 50 registros THEN o controle "Mostrar mais" SHALL não aparecer.

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
| --- | --- | --- | --- |
| DSP-01 | DSP-01: Estrutura de página | - | Pending |
| DSP-02 | DSP-02: Estados de tela | - | Pending |
| DSP-03 | DSP-03: Feedback e confirmação | - | Pending |
| DSP-04 | DSP-04: Formulários | - | Pending |
| DSP-05 | DSP-05: Listagens | - | Pending |
| DSP-06 | DSP-06: Permissões na interface | - | Pending |
| DSP-07 | DSP-07: Texto da interface | - | Pending |

**Coverage:** 7 total, 0 mapped to tasks, 7 unmapped.

---

## Success Criteria

- [ ] Todas as listagens da V1 abrem em 390 px sem rolagem horizontal de página, e só rolam na horizontal as tabelas com justificativa registrada.
- [ ] Um link copiado de uma listagem filtrada reproduz a mesma consulta.
- [ ] Nenhuma ação de linha repete nome acessível entre linhas.
