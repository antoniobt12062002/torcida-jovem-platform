# Web Design — STATE

Documento de coordenação da feature `web-design`: a linguagem visual e o Design System da Web (`web/`), aplicados ao que a Web V1 já entrega (Identity, Financeiro e Estoque). Decisões locais usam o prefixo **WDS-D-NNN**, que nunca reaproveita `AD-NNN` (`.specs/STATE.md`) nem `WEB-D-NNN` (`.specs/features/web/STATE.md`).

Esta feature **não altera a API** nem o comportamento funcional verificado da Web V1. Muda aparência, estrutura de tela, padrões de interação e acessibilidade. Onde uma decisão daqui revisa um requisito já verificado da Web V1, a revisão está listada em "Relação com decisões anteriores".

**Fase atual**: Specify/Discuss, Design e Tasks aprovados em 2026-10-11. Execute autorizado **somente para as rodadas 1 a 5** (até a apresentação do espécime). As rodadas 6 e 7 estão bloqueadas até aprovação explícita do mantenedor.

---

## Pergunta central

Como transformar a identidade da Torcida Jovem em um produto digital profissional, moderno, acessível e escalável, sem perder a personalidade e sem virar um "dashboard administrativo genérico com vermelho"?

## O que significa "Design System pronto"

1. A folha de tokens existe em três camadas, com a paleta oficial, e nenhum componente ou tela usa valor solto de cor.
2. A página de espécime mostra fundação, componentes, estados e exemplos responsivos, e foi aprovada pelo mantenedor antes da migração das telas.
3. O shell, o login e as páginas de erro carregam a identidade (preto, vermelho como acento, tipografia, escudo).
4. Todas as telas da Web V1 usam os mesmos padrões de página, lista, formulário, estado e feedback, e funcionam em 390, 820 e 1280 px.
5. Os critérios de aceite das seis sub-specs da Web V1 continuam valendo; os gates de `web` passam.
6. A revisão com `web-design-guidelines` não deixa achado sem correção ou sem conflito registrado, e a conferência visual no navegador foi executada e registrada.

---

## Mapa das sub-specs

| # | Spec | Requisitos | Conteúdo | Status |
| --- | --- | --- | --- | --- |
| 01 | `spec/01-fundacao-visual.md` | DSF-01 a DSF-06 | Tokens, cor, tipografia, espaço e forma, marca, espécime | Spec, design e tasks aprovados |
| 02 | `spec/02-componentes.md` | DSC-01 a DSC-07 | Kit de UI, componentes compostos e de domínio | Spec, design e tasks aprovados |
| 03 | `spec/03-padroes.md` | DSP-01 a DSP-07 | Página, estados, feedback, formulários, listagens, permissões, texto | Spec, design e tasks aprovados |
| 04 | `spec/04-shell-e-acesso.md` | DSS-01 a DSS-06 | Shell, navegação, telas de acesso, páginas de erro, metadados | Spec, design e tasks aprovados |
| 05 | `spec/05-migracao-web-v1.md` | DSM-01 a DSM-06 | Migração das telas existentes e verificação | Spec, design e tasks aprovados |

Cada sub-spec tem um documento de mesmo nome em `design/` e em `tasks/`. A direção visual aprovada está em `context.md` (saída do Discuss); os valores e as interfaces estão em `design/`. As unidades de execução, a propriedade de arquivos e a DAG estão em `design/05-migracao-web-v1.md` e resumidas abaixo.

Ordem de dependência prevista: 01 → 02 → (03 ∥ 04) → 05. O espécime (DSF-06) é o portão entre 02 e 05.

---

## Decisões (WDS-D-NNN)

Todas aprovadas pelo mantenedor em 2026-10-11, salvo indicação.

### WDS-D-001 — Paleta oficial (D1)
Preto `#111111`, vermelho `#D81E28`, branco `#FFFFFF`, aço `#A7A9AC`. Os tokens seguem esses valores. O escudo preserva a arte original, inclusive as cores do arquivo (`#E30D1E`, `#1B1B1B`, `#AD9F9C`), que não são ajustadas para coincidir com os tokens.

### WDS-D-002 — Papel do vermelho (D2)
Ação principal predominantemente em preto. Vermelho é identidade, destaque, estado ativo e ênfase, com peso visual e sem onipresença. Nem toda ação primária é vermelha.

### WDS-D-003 — Sem modo escuro neste ciclo (D3)
Tema único. A arquitetura de tokens permite um tema escuro futuro sem refazer componentes.

### WDS-D-004 — Tipografia (D4 e questão 1)
Barlow Condensed para títulos, identificação de seções e elementos de forte identidade. Geist para corpo, formulários, tabelas, botões e demais textos funcionais. Carregadas por `next/font`, sem dependência nova. Caixa alta com parcimônia. É uma decisão do Design System: não afirma que Barlow Condensed seja a fonte histórica da torcida.

### WDS-D-005 — Shell escuro e conteúdo claro (D5)
Navegação e moldura predominantemente escuras; conteúdo operacional claro.

### WDS-D-006 — Tabelas em telas estreitas (D6)
Abaixo de 768 px: lista ou cartões empilhados quando possível. De 768 a 1023 px: priorização de colunas. Desktop: tabela completa. Rolagem horizontal só quando realmente necessária. A solução não pode prejudicar tabelas pequenas só para eliminar toda rolagem.

### WDS-D-007 — Select (D7)
Select nativo estilizado em filtros e formulários simples. O componente do kit só quando houver busca, conteúdo rico ou interação mais complexa.

### WDS-D-008 — Escudo (D8 e questão 3)
O arquivo `web/public/brand/torcida-jovem-logo.svg` é o asset autorizado e pode ser versionado. Só o enquadramento do SVG é corrigido (remoção da área vazia da prancheta). O desenho não é alterado. Favicon reduzido, versão monocromática e versão negativa: sem asset oficial fornecido, ficam como dependência de design; nenhuma versão nova da marca é criada.

### WDS-D-009 — Página de espécime (D9)
Referência visual antes da migração das telas, com paleta, tokens, tipografia, hierarquia de títulos, botões, campos, badges e status, tabela, estados, shell e navegação e exemplos responsivos. Stitch não é usado neste ciclo sem autorização adicional.

### WDS-D-010 — Escopo (D10)
Fundação visual, tokens, componentes, padrões compartilhados, shell, login e migração das telas existentes da Web V1. Sem dashboards, áreas públicas ou módulos inexistentes criados para demonstrar capacidade.

### WDS-D-011 — Estado na URL (D11)
Filtros, busca, paginação, ordenação e aba ativa usam a URL quando isso melhora compartilhamento, navegação, recarga e os botões voltar e avançar. Só APIs nativas do Next e do React; sem `nuqs` nem outra biblioteca.

### WDS-D-012 — Listas longas (D12)
Sem virtualização e sem biblioteca nova. Sem paginação no servidor, usa-se "Mostrar mais" ou carregamento progressivo no cliente quando fizer sentido. Não é solução definitiva de desempenho: com volumes reais maiores, paginação ou virtualização é reavaliada em ciclo próprio.

### WDS-D-013 — A FERA e assets ausentes (questão 2)
Não há pacote de assets oficiais além do escudo. Nenhuma versão de A FERA é inventada ou gerada. Onde um asset seria necessário, registra-se a dependência (tabela abaixo).

### WDS-D-014 — Celular é uso real (questão 4)
Principalmente para operações administrativas rápidas, consulta, estoque e movimentações. Prioridades: alvos de toque adequados, leitura rápida, formulários simples, ações claramente identificadas, listas responsivas e baixa dependência de tabelas horizontais. Não se presume que as operações ocorram predominantemente em celular.

### WDS-D-015 — Personalidade (pares de contraste e orientação adicional)
Conteúdo claro com moldura escura; denso, mas agrupado e respirável; títulos fortes em página e seção, discretos no restante; predominantemente plano, com bordas e pouca sombra; convencional; sério, direto e funcional, com identidade própria. A identidade aparece por tipografia, proporção, uso do preto, vermelho como acento, escudo, iconografia, hierarquia, detalhes do shell, estados e componentes. Evitar: excesso de vermelho, gradientes decorativos, sombras excessivas, brilho ("glow"), cartões em excesso, estética genérica de SaaS e elementos de torcida caricatos na área operacional.

### WDS-D-016 — Skills como instrumento, projeto como autoridade
`frontend-blueprint` 1.0.0 alimenta Specify, Discuss e Design. `web-design-guidelines` 1.0.0 alimenta o diagnóstico e a Verification. O processo é o do `tlc-spec-driven`. Em conflito entre requisito aprovado, decisão em `.specs`, arquitetura existente e regra de skill, o projeto prevalece e o conflito é registrado (tabela abaixo).

### WDS-D-017 — Espécime em rota interna (P1)
Rota interna em `web/`, feita com os tokens e componentes reais, indisponível em produção. Não cria componente nem estilo exclusivo; exemplos de demonstração derivam dos componentes reais.

### WDS-D-018 — Verde e âmbar como cores de estado (P2)
Verde de sucesso e âmbar de aviso existem só como cores semânticas de estado. Não são identidade, não são ação primária, nunca comunicam sozinhas (sempre com texto, ícone ou símbolo) e precisam de contraste adequado nos tokens finais. Informação continua neutra, sem azul.

### WDS-D-019 — Navegação fixa a partir de 1024 px (P3)
Abaixo de 1024 px, navegação em gaveta; a partir de 1024 px, fixa. Substitui o comportamento de 768 px de FND-05 AC2 da Web V1. **Revalidação obrigatória**: o critério de aceite e o smoke responsivo são reexecutados em 390, 768, 820, 1024 px e desktop amplo; a verificação anterior não vale para o novo ponto de quebra.

### WDS-D-020 — Nome da marca na interface (P4)
"Torcida Jovem" no shell e na tela de login. "TJ Platform" só como título e metadado da aplicação, nunca como marca principal da interface.

### WDS-D-021 — Favicon (P5)
O escudo existente, redimensionado, sem redesenho e sem nova interpretação para ganhar legibilidade. Se não funcionar tecnicamente em algum contexto, registra-se a necessidade de asset oficial reduzido para ciclo posterior.

### WDS-D-022 — Radio, Switch, Tabs e Tooltip (P6)
Especificados no Design System, implementados no primeiro caso de uso real. Não entram na implementação deste ciclo se nenhuma tela da Web V1 os usar.

### WDS-D-023 — Registros fora desta feature (P7)
Nota de revisão de WEB-D-005 em `.specs/features/web/STATE.md` e `AD-018` em `.specs/STATE.md`, feitos em 2026-10-11. Mantém-se a distinção entre decisão de produto ou design, requisito, correção documental e decisão arquitetural; observação de implementação não vira decisão material.

### WDS-D-024 — Semântica de valores (P8)
Receita: verde semântico com sinal "+". Despesa: neutro com sinal "−". Saldo negativo: estado de erro com ícone e texto ou sinal. Formatação monetária, alinhamento e algarismos tabulares mantidos. Cor nunca é a única distinção.

### WDS-D-025 — Orientações do mantenedor para o Design (2026-10-11)
1. Barlow Condensed só fica em números de destaque se tiver boa leitura e tabulação; caso contrário, Geist assume valores financeiros e informação densa. 2. Tokens semânticos organizados por conceito (`brand`, `surface`, `content`, `state`, `focus`), sem acoplar componentes a hexadecimais. 3. Responsividade definida por componente e por padrão, não só "desktop e mobile". 4. Os achados da `web-design-guidelines` entram no Design System, e a skill é gate obrigatório do verificador final. 5. Nenhuma pendência de `design-pendencias.md` é corrigida antes do Design; as correções entram nas Tasks.

### WDS-D-026 — Decisões de design DD-1 a DD-6 (aprovadas em 2026-10-11)
- **DD-1**: nomes semânticos do kit mantidos, agrupados por conceito; sem renomeações desnecessárias; componentes não usam cor primitiva.
- **DD-2**: tabela e lista com semântica e acessibilidade adequadas; CSS sempre que viável; compatível com SSR e hidratação do Next; validação em 390, 768, 820, 1024 e 1440 px.
- **DD-3**: `CANCELADA` com tratamento neutro, ícone e texto; cancelamento não se confunde com erro.
- **DD-4**: saldo em Barlow Condensed, condicionado à avaliação visual do espécime; se a leitura de números financeiros for prejudicada, Geist.
- **DD-5**: ícone do iOS em PNG de 180 por 180 px, gerado do escudo oficial, sem redesenho nem modificação da arte.
- **DD-6**: estratégia de execução aprovada, com a tabela de propriedade substituída pela versão revisada (WDS-D-028).

### WDS-D-027 — Ajustes adicionais de design (2026-10-11)
Login em duas colunas no desktop: painel de marca com fundo preto, escudo oficial e a identificação "Torcida Jovem"; "Desde 2013" não é repetido como texto decorativo (a inscrição do escudo é preservada); no celular, coluna única. O tamanho visual de 24 px sai dos botões do kit; ícones podem ter 24 px, e as áreas interativas seguem os critérios de acessibilidade e toque das specs.

### WDS-D-028 — Unidades, propriedade e DAG (2026-10-11)
Doze unidades: DS-F, DS-T, DS-K, DS-L, DS-A, DS-E, DS-S, DS-ID, DS-FA, DS-FB, DS-ES e DS-INT, mais a verificação independente. DAG: rodada 1 DS-F; rodada 2 DS-K e DS-T em paralelo; rodada 3 DS-L; rodada 4 DS-A; rodada 5 DS-E, seguida de portão obrigatório de aprovação visual; rodada 6 DS-S, DS-ID, DS-FA, DS-FB e DS-ES em paralelo; rodada 7 DS-INT e verificação independente. A rodada 6 não começa automaticamente. DS-T é dona exclusiva de `test/msw/fixtures.ts`, `test/msw/server.ts`, `test/render.tsx`, `test/setup.ts`, `lib/session/navigation-mock.ts` e `lib/session/test-utils.tsx`. Depois do portão, os arquivos de DS-T, DS-K, DS-A e DS-L ficam congelados para as cinco frentes; mudança compartilhada passa por unidade de integração com responsável único; os alias de compatibilidade ficam até DS-INT, que os remove. Sem merge sem autorização explícita.

### WDS-D-029 — Graphify como ferramenta auxiliar (2026-10-11)
Usado para análise estrutural e planejamento da execução paralela, pelo executável local 0.9.86, com atualização manual pelo coordenador. Não é autoridade: specs e `STATE.md` prevalecem, seus resultados não viram requisito, e nenhuma decisão de paralelização depende só do grafo. Documentação operacional em `docs/development/ai-environment.md`; não há `AD` global para ele.

### WDS-D-030 — Ajuste de dois neutros na rodada 1 (2026-10-11)
Correção de implementação dentro da margem do design (valores finais dos neutros a critério do autor): o texto secundário passa de `#6B6E72` a `#64676B` e a borda de campo de `#8A8D91` a `#808387`. Motivo: sobre a superfície de hover (`accent`), os valores originais davam 4,37:1 em texto e 2,84:1 em borda, abaixo dos limites de DSF-02. Com os novos: texto 5,68 sobre branco, 5,26 sobre `muted`, 4,85 sobre `accent`; borda 3,81, 3,53 e 3,25. A paleta oficial não muda. Os pares entraram no teste de contraste.

### WDS-D-031 — Direção estética aprovada no portão do espécime e DD-4 definitiva (2026-10-12)
O mantenedor aprovou a direção visual do espécime: shell e sidebar pretos, escudo e tipografia condensada, vermelho pontual, Barlow Condensed nos títulos e Geist no conteúdo operacional, botões primários pretos, tabela no desktop e lista empilhada no celular, badges com significado textual, login em duas colunas no desktop e coluna única no celular, componentes, espaçamento e hierarquia. **DD-4 definitiva**: Barlow Condensed 700 nos saldos isolados de destaque; valores em tabelas, listas e formulários em Geist. Sem novas alterações estéticas além das correções pedidas. A liberação das rodadas 6 e 7 continua bloqueada até a correção do overflow em 1024 px, da barra de filtros e da conferência de toque, teclado e movimento reduzido no navegador.

### WDS-D-032 — Prioridade de colunas e barra de filtros pela largura do conteúdo (2026-10-12)
Correção pedida no portão do espécime: com a navegação fixa (a partir de 1024 px), a área de conteúdo em 1024 px tem cerca de 705 px úteis, e a listagem decidia pela largura da tela, gerando rolagem horizontal do documento. A listagem passa a ser um contêiner de consulta: as colunas de prioridade 3 só aparecem com tela de 1024 px ou mais **e** área de conteúdo de pelo menos 64rem (1024 px; inicialmente 56rem, elevado no smoke da rodada 6 porque valores na casa dos milhares ainda faziam a tabela de lançamentos rolar por dentro entre 1200 e 1265 px), e um invólucro contém qualquer sobra. Na mesma correção, "Forma de pagamento" dos lançamentos passou da prioridade 2 para a 3, porque as colunas de prioridade 1 e 2 não cabiam em 720 px. A barra de filtros passa a distribuir os campos pela largura do próprio contêiner (uma, duas ou várias colunas). O ponto de quebra da navegação não muda. DSP-05 AC5 foi ajustado para refletir isso.

### WDS-D-033 — Espécime aprovado e portão da rodada 5 encerrado (2026-10-12)
O mantenedor aprovou formalmente o espécime depois das correções de WDS-D-032 e autorizou a execução autônoma das rodadas 6 e 7 até a verificação independente, sem push, PR, deploy ou merge na `develop`.

### WDS-D-034 — Fechamento das lacunas da verificação, iteração 1 (2026-10-13)
- Os achados da `web-design-guidelines` sem registro foram corrigidos (`87fdd3a`, `63b5e8e`) e o mutante M13 foi morto (`ea87164`).
- `web/lib/forms/use-focus-invalid-after-submit.ts` (criado em `2903f32` para mover o foco ao primeiro campo inválido depois de um erro da API, N2 do smoke final) fica aceito em `lib/forms/`, ao lado de `applyProblemToForm`: é utilitário de formulário, sem regra de negócio; registro de propriedade feito aqui.
- DSM-01 AC2 foi ajustado para incluir `.specs/`, `docs/development/` e `.gitignore`, que mudaram com aprovação desde a base documental.
- **Lacunas de precisão da spec apontadas pelo verificador, aceitas como estão** (sem retrabalho neste ciclo): DSF-03 AC4 não lista o nome da marca em caixa alta (aprovado no espécime, WDS-D-031); DSF-04 AC8 fala em dois pontos de quebra, mas o diálogo usa 640 px (DSC-04 AC4) e o shell 400 px para esconder o nome; DSC-05 AC2 cita tom "informação", fundido ao neutro por WDS-D-018; DSC-01 AC1, DSC-05 AC5, DSP-02 AC6 e DSP-07 AC1-3 não têm valor verificável preciso; DSP-07 AC2 contra o link "Detalhes" (verbo implícito); DSC-04 AC5, DSP-01 AC3 e DSP-04 AC2 com evidência parcial. Ficam como entrada para a próxima revisão das specs.

---

## Relação com decisões anteriores

| Decisão anterior | Efeito |
| --- | --- |
| WEB-D-003 (bibliotecas) | Preservada. Nenhuma dependência nova neste ciclo. |
| WEB-D-005 (desktop-first; tabelas com rolagem horizontal ou listas empilhadas) | Revisada por WDS-D-006 e WDS-D-014: o desktop continua sendo o layout de referência, mas o celular passa a ter padrão definido, e a rolagem horizontal deixa de ser saída padrão. |
| WEB-D-012 (pt-BR) | Preservada. Frase comum, sem Title Case. |
| WEB-D-014 (autoria em UUID) e WEB-D-015 (saldo em lote) | Preservadas. O Design System define a apresentação do identificador; o dado continua sendo lacuna de API. |
| WEB-D-016 (listas sem paginação no servidor) | Preservada e complementada por WDS-D-012. |
| "Fora do V1": tema escolhido pelo usuário | Continua fora (WDS-D-003). |
| FND-05 AC2 (navegação recolhível abaixo do ponto de quebra de tablet) | Revisado por WDS-D-019: o ponto de quebra passa a 1024 px. O critério é reverificado neste ciclo (DSS-02). |
| FND-05 AC6 (confirmar desabilitado sem motivo) | Preservado, apesar da regra contrária da `web-design-guidelines`. |

Registrado em 2026-10-11 (WDS-D-023): nota de revisão em `.specs/features/web/STATE.md` e `AD-018` em `.specs/STATE.md`.

---

## Dependências de assets

| Asset | Situação | Tratamento |
| --- | --- | --- |
| Escudo colorido (`web/public/brand/torcida-jovem-logo.svg`) | Existe, ainda não versionado. Prancheta A4 (`viewBox 0 0 595.28 841.89`) com margens vazias; textos em curvas. | Versionar com o enquadramento corrigido (DSF-05). |
| Escudo reduzido para favicon | Não existe | **Necessário (constatado na rodada 1, 2026-10-11):** o escudo redimensionado é usado como ícone (WDS-D-021), mas em 16 px fica ilegível: só se reconhece o disco preto e vermelho; letreiro e ilustração viram mancha. Fica registrada a necessidade de um asset oficial reduzido para ciclo posterior. Nenhuma versão simplificada foi criada. |
| Escudo monocromático | Não existe | Dependência de design. Não é usado até existir. |
| Escudo em versão negativa | Não existe | Dependência de design. O escudo atual tem fundo preto próprio e funciona sobre o shell escuro e sobre fundo claro. |
| A FERA (mascote) | Não existe | Dependência de design. Não aparece em nenhuma tela até existir asset oficial. |
| Manual de marca e artes da torcida | Não existem no repositório | A direção visual se apoia no escudo e nas decisões WDS-D. |
| Fontes | Barlow Condensed e Geist, licença OFL, via `next/font` | Sem arquivo novo no repositório. |

---

## Conflitos entre skills e projeto

| Regra da skill | Decisão do projeto | Resultado |
| --- | --- | --- |
| Title Case em títulos e botões (`web-design-guidelines`) | pt-BR em frase comum (WEB-D-012) | Regra não se aplica. |
| Botão de envio habilitado até a requisição começar (`web-design-guidelines`) | Confirmar desabilitado sem motivo ou quantidade (FND-05 AC6) | Comportamento atual mantido. |
| Estado na URL com `nuqs` (`web-design-guidelines`) | Sem dependência nova | APIs nativas (WDS-D-011). |
| Virtualização acima de 50 itens com `virtua` (`web-design-guidelines`) | Sem dependência nova | "Mostrar mais" (WDS-D-012). |
| Regras de dark mode (`web-design-guidelines`) | Sem modo escuro (WDS-D-003) | Valem só `theme-color` e `color-scheme` do tema único. |
| Prototipagem no Stitch (`frontend-blueprint`) | Sem serviço externo sem autorização | Fase substituída pelo espécime (WDS-D-009). |
| Plano de execução e construção atômica (`frontend-blueprint`) | Processo do `tlc-spec-driven` | Substituídos por Design, Tasks e Execute. |
| Modelo "data-dense" com base em tema escuro (`frontend-blueprint`) | Shell escuro, conteúdo claro (WDS-D-005) | Só a moldura é escura. |
| Alvo de toque de 44 px em todo lugar (`frontend-blueprint`) | Densidade de desktop | 44 px em ponteiro de toque; controles de 36 px em mouse (DSC-01). |

---

## Varredura de dimensões implícitas (escopo Large)

| Dimensão | Resultado |
| --- | --- |
| Validação e limites de entrada | DSP-04 (formulários) e DSP-05 (parâmetros de URL inválidos são ignorados). |
| Falha e falha parcial | DSP-02, DSP-03 e DSS-04. |
| Idempotência, repetição, duplicidade | N/A: nenhuma escrita nova; as escritas existentes não mudam. |
| Fronteiras de autenticação e limites | Sem mudança; DSP-06 preserva FND-04. O espécime não fica disponível em produção (DSF-06). |
| Concorrência e ordenação | N/A: sem estado compartilhado novo. |
| Ciclo de vida de dados | N/A: o único estado novo é o da URL; nada é gravado no navegador. |
| Observabilidade | N/A: sem registro novo; as páginas de erro não exibem detalhe técnico (DSS-04). |
| Falha de dependência externa | Fontes são servidas pela própria aplicação (`next/font`); sem chamada a terceiros em tempo de execução. |
| Integridade de transição de estado | N/A: os fluxos de estado (lançamento, usuário) não mudam; só a apresentação do status (DSC-05). |

---

## Autorizações de execução

| # | Ponto | Situação |
| --- | --- | --- |
| E-1 | Tasks (44), DAG de sete rodadas, propriedade, gates e verificação independente | Aprovados em 2026-10-11 |
| E-2 | Commit da base documental na branch `feature/web-design` | Autorizado em 2026-10-11; sem push e sem PR |
| E-3 | Worktrees por unidade, subagentes e Playwright MCP (só ambiente local e dados de teste) | Aprovados em 2026-10-11 |
| E-4 | Rodadas 1 a 5 | Autorizadas, com commits locais nas branches das unidades e integração local na `feature/web-design` com gates verdes |
| E-5 | Rodadas 6 e 7 | **Autorizadas em 2026-10-12** após a aprovação formal do espécime; rodada 7 só com todos os gates da rodada 6 verdes |
| E-6 | Merge na `develop`, push e PR | Não autorizados |

## Mapa das unidades de execução

| Rodada | Unidade | Tasks | Arquivo |
| --- | --- | --- | --- |
| 1 | DS-F Fundação visual | T1 a T4 | `tasks/01-fundacao-visual.md` |
| 2 | DS-T Infraestrutura de testes | T1 e T2 | `tasks/02-componentes.md` |
| 2 | DS-K Kit de componentes | T3 a T8 | `tasks/02-componentes.md` |
| 3 | DS-L Bibliotecas auxiliares | T1 e T2 | `tasks/03-padroes.md` |
| 4 | DS-A Componentes compostos | T3 a T10 | `tasks/03-padroes.md` |
| 5 | DS-E Espécime visual e portão | T5 a T7 | `tasks/01-fundacao-visual.md` |
| 6 | DS-S Shell e acesso | T1 a T5 | `tasks/04-shell-e-acesso.md` |
| 6 | DS-ID Identity | T1 e T2 | `tasks/05-migracao-web-v1.md` |
| 6 | DS-FA Financeiro A | T3 e T4 | `tasks/05-migracao-web-v1.md` |
| 6 | DS-FB Financeiro B | T5 a T7 | `tasks/05-migracao-web-v1.md` |
| 6 | DS-ES Estoque | T8 a T10 | `tasks/05-migracao-web-v1.md` |
| 7 | DS-INT Integração | T11 a T13 | `tasks/05-migracao-web-v1.md` |
| 7 | VERIFY | T14 | `tasks/05-migracao-web-v1.md` |

A numeração `T1…Tn` é local a cada arquivo, como na Web V1. Total: 44 tasks.

## Fora deste ciclo

| Item | Motivo |
| --- | --- |
| Modo escuro e seletor de tema | WDS-D-003. |
| Dashboards, indicadores e gráficos | WDS-D-010; a API não tem endpoint de indicadores. |
| Áreas públicas, loja, área do associado | WDS-D-010; módulos inexistentes. |
| Virtualização e paginação no servidor | WDS-D-012. |
| Nomes de autores e número legível de lançamento | Exigem mudança na API (WEB-D-014). |
| Regra de negócio de ações em registros inativos ou cancelados | Decisão de produto dos módulos; o Design System só define como mostrar ação indisponível. |
| Aviso de alterações não salvas ao sair de formulário | Ideia registrada em `context.md`; os formulários da V1 são curtos e em diálogo. |
| Storybook ou ferramenta de documentação de componentes | Dependência nova; o espécime cumpre o papel. |

---

## Registro de execução

| Rodada | Unidade | Commits | Testes ao fim (arquivos / testes) | Observações |
| --- | --- | --- | --- | --- |
| Base | — | `7c30c6a` | 69 / 851 | Base documental |
| 1 | DS-F | `ea86873`, `2a7a4ec`, `ca30de7`, `772b97e`, `a726389` | 74 / 969 | Ajuste de neutros (WDS-D-030); escudo ilegível em 16 px |
| 2 | DS-T | `c82cc49`, `20f2609` | 76 / 1005 | Os 33 consumidores do mock de navegação passam sem alteração |
| 2 | DS-K | `ae398cc`, `cbef5e2`, `c5fdf35`, `845a107`, `316eb3d`, `063b294`; integração `08c971d` | 79 / 1207 (com DS-T) | Alias de compatibilidade listados abaixo |

| 3 | DS-L | `dd46c28`, `5bb3295` | 83 / 1285 | Estado na URL e consulta de mídia com teste de SSR e hidratação |
| 4 | DS-A | `f74768f`, `82f6dc6`, `aa868d8`, `ea9f5e6`, `fb5500d`, `8577209`, `e9aeeb5`, `0dd89cd` | 93 / 1493 | Guarda de `components/app` sem exceções |
| 4 | Correção do gate | `7476156`, `8c20dfd` | 93 / 1517 | Revisão `web-design-guidelines` reprovou 6 pontos (foco de item de menu e select, progresso na confirmação, HTML inválido na marca, alvos da trilha e da marca do shell); corrigidos com mais 5 melhorias de baixo risco |

| 5 | DS-E | `7f72ba6`, `d529b4f`; evidência neste commit | 97 / 1612 | Espécime em `/design-system`, 404 em produção; aguardando aprovação visual |

| 6 | DS-FA, DS-FB, DS-ES, DS-ID, DS-S | integrações `a629608` e seguintes; ícones de navegação `fc73eac`; correções do smoke `7ebae02`, `32184e8`, `c0b1925` | 107 / 1761 | Smoke real com API: 5 fluxos passaram; 4 defeitos corrigidos; FND-05 AC2 reverificado |

**Alias de compatibilidade a remover em DS-INT** (marcados no código com "Alias de compatibilidade: removido em DS-INT."): `button.tsx` (variantes `default` e `outline`; tamanhos `default`, `icon`, `icon-sm`), `badge.tsx` (`default`, `secondary`, `destructive`), `alert.tsx` (`default` e `destructive`), `table.tsx` (`scroll="legacy"` como padrão).

**Conflitos com a `web-design-guidelines` registrados na rodada 4** (o projeto prevalece; reavaliar na rodada 7):

- Áreas seguras: `viewport` sem `viewportFit: "cover"`, então o navegador mantém a página dentro da área segura e os `env(safe-area-inset-*)` valem zero; DSS-02 AC6 fica atendido por omissão. Ativar `cover` exigiria insets na gaveta esquerda, no topo do diálogo em tela cheia e no layout de acesso.
- Placeholder `0,00` do campo de dinheiro sem "…": é o exemplo de formato de FND-05 AC4, mantido.
- O toast anima `height` além de `transform` e `opacity`: é o empilhamento do Base UI.
- Datas de calendário formatadas por expressão regular, sem `Intl`: de propósito, para não aplicar fuso a uma data sem hora.
- O anel de foco dos itens de menu também aparece no hover do ponteiro, porque o Base UI marca `data-highlighted` nos dois casos.

**Pendências registradas para a rodada 7** (melhorias B da revisão, não bloqueiam): `Button` em `loading` perde o foco do teclado; rótulo longo de botão não quebra em 320 px; `Checkbox` solto abaixo de 44 px; `Table` sem nome acessível obrigatório no tipo; `CardTitle` com nível fixo; link de pular ausente durante a troca obrigatória de senha; "Início" e módulos sem ícone; `RowActions` sem motivo para ação indisponível; campo de motivo sem `name`; foco do campo de dinheiro sem `focus-within`; escrita pendente da busca descartada ao sair da tela; `tap-highlight` e `touch-action` na camada base.

**Registrado em DS-INT (T11, 2026-10-12):**

- Alias removidos de `button.tsx`, `badge.tsx`, `alert.tsx` (o tom do `Alert` passou a ser obrigatório), `table.tsx` (sem `scroll`, a tabela não tem contêiner de rolagem; `scrollable` só com justificativa) e `status-badge.tsx` (só `StatusDef`). Nenhum uso restava fora dos testes do kit. Guarda global em `web/test/design/global.guard.test.ts`; testes alterados no ciclo em `evidence/testes-alterados.md`.
- **D5, limitação da biblioteca**: no `@base-ui/react` 1.8.0 instalado, `Toast.Root` fixa `role="dialog"` (ou `alertdialog` na prioridade alta), `aria-modal=false` e `tabIndex=0` nas propriedades padrão (`toast/root/ToastRoot.js`). A API tipada (`ToastRootProps`) não oferece opção para trocar o papel nem o foco; sobrescrever por atributo solto dependeria da ordem interna de mescla e romperia o modelo de teclado da biblioteca (F6 leva à região "Avisos" e Tab percorre os avisos até "Fechar" e a ação). O anúncio não depende desse papel: a região viva própria do Base UI (`role="alert"` no viewport) faz o anúncio. Mantido como está.
- `FilterBar` sem filtros além da busca (estoque passa `{null}`): em tela estreita não mostra o botão "Filtros" (a gaveta ficaria vazia); "Limpar filtros" aparece junto da busca preenchida.

**Desvios aceitos pelo coordenador na rodada 2, a rever em DS-INT ou no espécime:**

- O toast de erro é anunciado como `polite`, não `assertive`: no Base UI a prioridade alta duplica o texto do aviso e quebraria 9 testes existentes de `features/` e `lib/session`, que estão congelados. Reavaliar em DS-INT, junto com esses testes.
- O `AlertDialog` continua centralizado abaixo de 640 px (DSC-04 AC4 cita só o `Dialog`).
- Espaçamento de letras dos títulos usa a escala padrão (`tracking-normal` e `tracking-wider`), não 0,01em e 0,06em.
- Área segura do dispositivo aplicada por estilo em linha em toast, rodapé de diálogo e gaveta inferior; não é verificável em jsdom e fica para a conferência no navegador.
- `Sheet` aceita só os lados `left` e `bottom`.

## Handoff

- **Fase**: Execute. Rodadas 1 a 6 concluídas e integradas, gates verdes (107 arquivos, 1761 testes, build, contrato, audit) e smoke real da rodada 6 aprovado (`evidence/smoke-rodada6.md`). Rodada 7 liberada.
- **Pendências para a rodada 7**: remoção de alias; D5 (toast como `role="dialog"`); percurso completo só por teclado e zoom de 200% (DSM-06 AC4 e AC5) no smoke final; item de T2 da DS-S "exceção forçada mostra a página de erro dentro do shell" (coberto por teste de componente; sem conferência no navegador).
- **Bloqueios**: nenhum. Sem push, PR ou merge na `develop`.
