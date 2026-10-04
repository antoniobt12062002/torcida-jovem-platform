# Validação independente — `estoque` V1

Verificador independente (não é o autor do código). Regra: sem evidência, a nota é zero. Data: 2026-10-04.

## Veredito

**Result**: PASS

**PASS, após a re-verificação 1 (`ab7eae9`).** A seção "Re-verificação 1 (ab7eae9)", no fim deste documento, traz a evidência. Nessa rodada os 2 mutantes que tinham sobrevivido foram mortos e os 5 ACs fracos ou parciais passaram a ter teste que afirma o que a spec define. Ficam itens abertos para o mantenedor, mas são de redação da spec e não bloqueiam o PASS.

O texto abaixo, até a seção da re-verificação, é a primeira passada sobre `17ba763`. Ele foi mantido como histórico.

**Veredito da primeira passada: FAIL.**

Os gates estão verdes e 16 dos 18 mutantes foram mortos. Mesmo assim o veredito é FAIL, por dois motivos:

- **2 mutantes sobreviveram (M5 e M18).** Cada um fica numa regra que a spec exige.
- **2 ACs têm uma subcláusula que nenhum teste cobre:** MOV-02 AC5, no nível do caso de uso, e MOV-03 AC2, no caso "outra `DEVOLUCAO`".

Nenhum desses achados é regressão que apareça em produção hoje: o código atual se comporta como a spec pede. O problema é que os testes não impedem que essas regras quebrem no futuro.

## Diff range

`9b8c65699b3de2f47ee89d78d2389209ffeee272..17ba7639d55c85d2577e3bff5dbe9533f9c2b8e1`: 22 commits, 72 arquivos, +8841/−38.

## Evidência por AC

Legenda da coluna "Bate com a spec?":

- **Sim**: o teste afirma o valor que a spec define.
- **Fraco**: o teste afirma menos do que a spec define.
- **Parcial**: uma subcláusula do AC fica sem teste.
- **Inspeção**: o AC é estrutural e não tem como ser testado por comportamento.

Caminhos relativos a `api/internal/estoque/`, salvo indicação.

| AC | Teste (`file:line`) | Valor afirmado | Bate com a spec? |
| --- | --- | --- | --- |
| PRD-01 AC1 | `app/criar_produto_test.go:16`; `http/produtos_handler_test.go:17` | `ID != ""` e campos iguais à entrada; HTTP `201` com os campos | Sim |
| PRD-01 AC2 | `app/criar_produto_test.go:32`; `http/produtos_handler_test.go:48`; `httpapi/estoque_e2e_test.go:23-24` | `ErrCodigoDuplicado`; `409 codigo_duplicado` | Sim |
| PRD-01 AC3 | `app/criar_produto_test.go:51` (3 campos); `http/produtos_handler_test.go:36` (só espaços); `httpapi/estoque_e2e_test.go:44-46` (vazio) | `ErrCampoObrigatorio`; `422 campo_obrigatorio` (espaços) e `422 validation_failed` (vazio) | Sim, mas a spec diz só "um erro de validação" (ver lacuna 6) |
| PRD-01 AC4 | `app/criar_produto_test.go:69`; `http/produtos_handler_test.go:87` | `ErrForbidden` / `403 forbidden` | **Fraco**: usa um `codigo` que não existe ("X"), então não prova que a resposta não revela um código existente. O código cumpre por ordem: `Authz.Require` vem antes de qualquer leitura (`app/criar_produto.go:47-52`) |
| PRD-02 AC1 | `app/listar_produtos_test.go:15`; `http/produtos_handler_test.go:63`; e2e `:26-30` | `len == 3`; o id criado aparece na lista; `len == 1` | **Fraco**: a spec pede "exatamente como criados", mas nenhum teste compara os campos listados com os criados |
| PRD-02 AC2 | `app/listar_produtos_test.go:37`; `http/produtos_handler_test.go:87` (caso "list") | `forbidden` / `403` | Sim |
| MOV-01 AC1 | `app/registrar_movimentacao_test.go:18`; `http/movimentacoes_handler_test.go:30` | `Quantidade == 10` (positiva), `Tipo == ENTRADA`; `201`, `quantidade == 10` | Sim |
| MOV-01 AC2 | `app/registrar_movimentacao_test.go:36`; `http/movimentacoes_handler_test.go:108` | `ErrProdutoNaoEncontrado`; `404 produto_nao_encontrado` | Sim |
| MOV-01 AC3 | `app/registrar_movimentacao_test.go:50` (0 e −1); `http/movimentacoes_handler_test.go:95` | `ErrQuantidadeInvalida`; `422 quantidade_invalida` | Sim |
| MOV-01 AC4 | `app/registrar_movimentacao_test.go:66`; `http/movimentacoes_handler_test.go:140` | `ErrForbidden`; `403 forbidden` (DIRETORIA) | Sim |
| MOV-02 AC1 | `app/registrar_movimentacao_test.go:80`; `httpapi/estoque_e2e_test.go:66-72` | `Quantidade == -4`; `201`, `quantidade == -2` | Sim |
| MOV-02 AC2 | `app/registrar_movimentacao_test.go:103` (inclui "nada registrado", `:120-123`); `http/movimentacoes_handler_test.go:47`; e2e `:62-64` e `:80-82` | `ErrSaldoInsuficiente` com 1 movimentação; `409 saldo_insuficiente` | Sim |
| MOV-02 AC3 | `app/registrar_movimentacao_test.go:36` (só ENTRADA) | `ErrProdutoNaoEncontrado` | Sim por caminho compartilhado: `Produtos.Buscar` vem antes do `switch` (`app/registrar_movimentacao.go:92`) |
| MOV-02 AC4 | `app/registrar_movimentacao_test.go:50` (só ENTRADA) | `ErrQuantidadeInvalida` | Sim por caminho compartilhado (`app/registrar_movimentacao.go:89`) |
| MOV-02 AC5 | `infra/movimentacao_repository_test.go:160` (concorrência), `:233` (serialização determinística) | 1 sucesso e 1 `saldo_insuficiente`; saldo final 0; B bloqueia enquanto A segura o lock | **Parcial**: os dois testes reimplementam a checagem dentro do próprio teste (`:172-184`) e nunca chamam `RegistrarMovimentacao.Execute`. A cláusula "via lock na mesma transação" do caso de uso não tem teste: o mutante **M18 sobreviveu** |
| MOV-03 AC1 | `app/registrar_movimentacao_test.go:127` (devolução de SAIDA → `+2`, soma 8) e `:166` (devolução de ENTRADA → `-2`); `http/movimentacoes_handler_test.go:60`; e2e `:74-76` | sinal invertido conforme a referência (EST-D-008) | Sim. O teste HTTP `:60` não confere o sinal, mas o teste de app confere |
| MOV-03 AC2 | `app/registrar_movimentacao_test.go:206` (inexistente, produto diferente, AJUSTE); `http/movimentacoes_handler_test.go:82` (sem referência) | `ErrDevolucaoInvalida`; `422 devolucao_invalida` | **Parcial**: o caso "referencia outra `DEVOLUCAO`" não tem teste em lugar nenhum. O mutante **M5 sobreviveu** |
| MOV-03 AC3 | `app/registrar_movimentacao_test.go:166-189` | saldo 2, devolver 3 da ENTRADA → `ErrSaldoInsuficiente` | Sim |
| MOV-03 AC4 | `app/registrar_movimentacao_test.go:255` | `original.Quantidade == saida.Quantidade` depois da devolução | Sim. Reforçado pelo banco: `GRANT INSERT, SELECT` sem `UPDATE` (`migrations/000009_estoque_movimentacoes.up.sql`) |
| MOV-04 AC1 | `app/listar_movimentacoes_test.go:17`; `infra/movimentacao_repository_test.go:79`; `http/movimentacoes_handler_test.go:120` | 3 itens em ordem `[entrada, saida, ajuste]` por ID | Sim (o teste HTTP só confere `len == 2`) |
| MOV-04 AC2 | `app/listar_movimentacoes_test.go:55`; `http/movimentacoes_handler_test.go:140` (caso "list") | `forbidden` / `403` | Sim |
| AJS-01 AC1 | `app/ajustar_estoque_test.go:17`; `http/ajustes_saldo_handler_test.go:18`; e2e `:104-112` | `Quantidade == -3`, `Tipo == AJUSTE`, motivo gravado, soma −3; `201`, saldo −3; e2e saldo −2 | Sim |
| AJS-01 AC2 | `app/ajustar_estoque_test.go:44` (`"   "`, nada persistido); `http/ajustes_saldo_handler_test.go:41` (`""`); e2e `:101-103` | `ErrMotivoObrigatorio`; `422 motivo_obrigatorio` | Sim |
| AJS-01 AC3 | `app/ajustar_estoque_test.go:63`; `http/ajustes_saldo_handler_test.go:54` | `ErrQuantidadeInvalida`; `422 quantidade_invalida` | Sim |
| AJS-01 AC4 | `app/ajustar_estoque_test.go:78`; `http/ajustes_saldo_handler_test.go:67` | `ErrProdutoNaoEncontrado`; `404` | Sim |
| AJS-01 AC5 | `app/ajustar_estoque_test.go:92` (só `movimentacao:create`); `http/ajustes_saldo_handler_test.go:80` (DIRETORIA) | `ErrForbidden`; `403 forbidden` | Sim |
| AJS-02 AC1 | `app/consultar_saldo_test.go:16` (0 e −5); `http/ajustes_saldo_handler_test.go:94` (7); `infra/saldo_test.go:32` (4 tipos → 6); e2e `:92-96` e `:108-112` | valores exatos de saldo, zero e negativo incluídos | Sim |
| AJS-02 AC2 | `app/consultar_saldo_test.go:43`; `http/ajustes_saldo_handler_test.go:127` | `ErrProdutoNaoEncontrado`; `404` | Sim |
| AJS-02 AC3 | `app/consultar_saldo_test.go:57` (só `movimentacao:read`); `http/ajustes_saldo_handler_test.go:113` | `ErrForbidden`; `403` | Sim no app. O teste HTTP diz "only movimentacao:read" no nome, mas usa ASSOCIADO, que não tem nenhuma permissão (lacuna 7) |
| PERM-01 AC1 | `module_test.go:32` | 6 permissões, sem duplicata, nomes exatos | Sim |
| PERM-01 AC2 | `module_test.go:60`, `:71`, `:83`, `:108`, `:131`; `identity/infra/role_repository_test.go:403-414` (matriz de produção); `cmd/bootstrap-admin/bootstrap_integration_test.go:53` | ESTOQUE_LOJA = 6, DIRETORIA = CONSELHO_FISCAL = 3 de leitura, nenhum outro papel, PRESIDENTE automático; 6 sincronizadas | Sim |
| PERM-01 AC3 | Inspeção: `module.go:80-93` usa só `estoqueapp.Perm*` | constantes, nenhum literal | Inspeção. O AC é estrutural e o compilador garante |
| API-01 AC1-3 | `http/produtos_handler_test.go:17`, `:63`, `:87`; `httpapi/estoque_e2e_test.go:16` | `201`, `200` com o item, `403 forbidden` | Sim (para o AC2, vale a ressalva de PRD-02) |
| API-02 AC1 | `http/movimentacoes_handler_test.go:30`, `:60`; e2e `:54` | `201` | Sim |
| API-02 AC2 | `http/movimentacoes_handler_test.go:120`; e2e `:78-82` | `200`, `len` 2 / 3 | Sim. A ordem é provada no app e na infra |
| API-02 AC3 | `http/movimentacoes_handler_test.go:47`; e2e `:62-64` | `409 saldo_insuficiente` | Sim. Devolução de ENTRADA via HTTP não tem teste; o mapeamento é o mesmo |
| API-02 AC4 | `http/movimentacoes_handler_test.go:108` | `404 produto_nao_encontrado` | Sim |
| API-02 AC5 | `http/movimentacoes_handler_test.go:82` (só "sem referência"); `http/handler_internal_test.go:34` | `422 devolucao_invalida` | **Fraco**: as 3 variantes que a spec cita (não existe, outro SKU, AJUSTE/DEVOLUCAO) não são exercitadas via HTTP. O mapeamento é coberto pelo teste de tabela |
| API-03 AC1-3 | `http/ajustes_saldo_handler_test.go:18`, `:41`, `:94`; e2e `:87` | `201` com saldo negativo; `422 motivo_obrigatorio`; `200 {saldo}` exato | Sim |
| API-04 AC1-2 | `http/handler_internal_test.go:25-61` (8 sentinels + 2 transversais, com e sem wrap, `application/problem+json`); `:66` (code ≠ `.Error()`) | status e `code` exatos da tabela EST-D-006 | Sim |

**Contagem:** 45 ACs.

| Situação | Qtd. | ACs |
| --- | --- | --- |
| Verificados com teste que bate com a spec | 39 | — |
| Verificado por inspeção | 1 | PERM-01 AC3 |
| Fracos | 3 | PRD-01 AC4, PRD-02 AC1, API-02 AC5 |
| Parciais, com subcláusula sem teste | 2 | MOV-02 AC5, MOV-03 AC2 |
| Sem nenhum teste | 0 | — |

## Sensor de discriminação

Os mutantes rodaram numa worktree destacada fora do repositório (`git worktree add --detach <scratchpad>/wt-mut HEAD`). O baseline sem mutação na worktree passou: `estoque/...` e `httpapi/...` com `-tags=integration -p 1`, tudo `ok`. Cada mutante foi aplicado como substituição única, conferida com contagem igual a 1. Depois disso: `go build ./...` (todos com saída 0), os pacotes-alvo com `go test -tags=integration -count=1` e o arquivo restaurado.

| # | Mutação | `file:line` | Resultado |
| --- | --- | --- | --- |
| M1 | `if saldo+qtd < 0` → `if false && saldo+qtd < 0` (guarda removida) | `app/registrar_movimentacao.go:134` | **Morto**: `TestRegistrarMovimentacaoSaidaWithInsufficientBalanceIsRejected`, `…DevolucaoOfAnEntradaRemovesEstoqueAndCanBeRejectedByBalance`, `TestCreateMovimentacaoSaidaAboveBalanceAnswers409`, `TestEstoqueE2EMovimentacoesFlow` |
| M2 | `saldo+qtd < 0` → `saldo+qtd < -1` (off-by-one) | `app/registrar_movimentacao.go:134` | **Morto**: os 2 testes de app de M1 e `TestEstoqueE2EMovimentacoesFlow` (o pacote HTTP não mata) |
| M3 | `if ref.Tipo == domain.Saida` → `== domain.Entrada` (inverte o sinal da devolução, EST-D-008) | `app/registrar_movimentacao.go:117` | **Morto**: `TestRegistrarMovimentacaoDevolucaoOfASaidaRepõeEstoque`, `…DevolucaoOfAnEntrada…` |
| M4 | remove `ref.ProdutoID != in.ProdutoID \|\|` (EST-D-009, outro SKU) | `app/registrar_movimentacao.go:114` | **Morto**: `TestRegistrarMovimentacaoDevolucaoWithAnInvalidReferenceIsRejected` |
| M5 | `(ref.Tipo != domain.Entrada && ref.Tipo != domain.Saida)` → `(ref.Tipo == domain.Ajuste)` (aceita referenciar outra DEVOLUCAO, EST-D-009) | `app/registrar_movimentacao.go:114` | **SOBREVIVEU**: app, http e httpapi todos `ok` |
| M6 | a mesma expressão → `(ref.Tipo == domain.Devolucao)` (aceita referenciar AJUSTE) | `app/registrar_movimentacao.go:114` | **Morto**: `TestRegistrarMovimentacaoDevolucaoWithAnInvalidReferenceIsRejected` |
| M7 | `if motivoBlank(in.Motivo)` → `if false && …` (motivo deixa de ser obrigatório) | `app/ajustar_estoque.go:60` | **Morto**: `TestAjustarEstoqueWithBlankMotivoIsRejectedBeforeAnyPersistence`, `TestCreateAjusteWithoutMotivoAnswers422`, `TestEstoqueE2EAjusteESaldoFlow` |
| M8 | `motivoBlank(in.Motivo)` → `in.Motivo == ""` (aceita motivo só com espaços) | `app/ajustar_estoque.go:60` | **Morto**: `TestAjustarEstoqueWithBlankMotivoIsRejectedBeforeAnyPersistence` |
| M9 | `RoleConselhoFiscal: reads` → `operationalPerms` (concede create/adjust ao CF) | `module.go:101` | **Morto**: `TestContributionGrantsConselhoFiscalExactlyTheThreeReadPermissions`, `TestContributionNeverGrantsConselhoFiscalCreateOrAdjust` e mais 2 |
| M10 | remove `PermSaldoRead` de `reads` (DIRETORIA e CF perdem `saldo:read`) | `module.go:92` | **Morto**: `TestContributionGrantsDiretoriaExactlyTheThreeReadPermissions`, `…ConselhoFiscal…` |
| M11 | `ErrSaldoInsuficiente` `409` → `422` | `http/handler.go:114` | **Morto**: `TestEveryDomainErrorMapsToItsStatusAndCode`, `TestCreateMovimentacaoSaidaAboveBalanceAnswers409` |
| M12 | code `devolucao_invalida` → `devolucao_invalid` | `http/handler.go:111` | **Morto**: `TestEveryDomainErrorMapsToItsStatusAndCode`, `TestCreateMovimentacaoDevolucaoWithoutReferenceAnswers422` |
| M13 | `if qtd < 0` → `if qtd < 0 && in.Tipo == domain.Saida` (devolução de ENTRADA deixa de ser checada) | `app/registrar_movimentacao.go:129` | **Morto**: `…DevolucaoOfAnEntradaRemovesEstoqueAndCanBeRejectedByBalance` |
| M14 | `ConsultarSaldo` ignora o erro de `Produtos.Buscar` | `app/consultar_saldo.go:42` | **Morto**: `TestConsultarSaldoWithAMissingProdutoIsRejected`, `TestGetSaldoForAnUnknownProdutoAnswers404` |
| M15 | `ORDER BY criado_em` → `ORDER BY criado_em DESC` | `infra/movimentacao_repository.go:74` | **Morto**: `TestListarMovimentacoesReturnsEveryTipoInCreationOrder`, `TestListarPorProdutoReturnsInCreationOrder` |
| M16 | `pg_advisory_xact_lock(hashtext(?))` → `SELECT 1` (lock removido) | `infra/movimentacao_repository.go:97` | **Morto**, mas só por `TestSaldoComLockSerializesConcurrentCallersOfTheSameProduto`. O teste de concorrência `:160` não mata sozinho |
| M17 | `CriarProduto` exige `PermProdutoRead` em vez de `PermProdutoCreate` | `app/criar_produto.go:47` | **Morto**: 21 testes de app e `TestProdutoOperationsWithoutThePermissionAre403Forbidden` |
| M18 | a checagem `SaldoComLock` + `saldo+qtd < 0` sai de dentro de `r.Tx(...)` e passa a rodar antes da transação. O advisory lock vira autocommit e é liberado na hora, e a checagem deixa de ser atômica com o `Criar` | `app/registrar_movimentacao.go:127-137` | **SOBREVIVEU**: app, http e httpapi todos `ok` |

**Resultado:** 18 mutantes, 16 mortos e 2 sobreviventes (M5 e M18).

Observação sobre M16: na primeira tentativa a substituição foi `SELECT 1 WHERE ? IS NOT NULL`. Ela falhava por erro de SQL, o que é uma morte inválida, então foi refeita como `SELECT 1`. Só o resultado refeito conta.

## Gates

Todos rodaram na árvore real, em `api/`, com HEAD `17ba763`.

| Gate | Resultado | Evidência |
| --- | --- | --- |
| `go vet ./...` | PASS | saída 0 |
| `go build ./...` | PASS | saída 0 |
| `go test ./...` | PASS | todos os pacotes `ok`, incluindo `internal/estoque` e `internal/estoque/http` |
| `golangci-lint run --path-mode=abs --build-tags=integration` | PASS | `0 issues.` (v2.14.0) |
| `go test -tags=integration ./...` | PASS, após nova execução (ver nota) | ver nota |

Nota sobre o gate de integração:

1. **Primeira execução:** falhou por infraestrutura, com `reaper: … Conflict` do testcontainers.
2. **Segunda execução** (comando idêntico): 4 pacotes falharam: `cmd/api`, `financeiro/http`, `httpapi` e `platform/audit`. Todas as 81 linhas de falha eram `falha ao iniciar via Docker` (daemon saturado criando containers em paralelo). Nenhuma era assertiva. Todos os pacotes de estoque passaram nessa execução: `estoque` `ok`, `estoque/app` `ok`, `estoque/http` `ok` e `estoque/infra` `ok`.
3. **Nova execução dos 4 pacotes**, em série com `-p 1`: `ok` em todos (`cmd/api` 6.9s, `financeiro/http` 9.2s, `httpapi` 14.2s, `platform/audit` 5.9s).

## Lacunas e achados

Ordenados do mais grave para o menos grave.

1. **M18 sobreviveu (MOV-02 AC5).** Nenhum teste exercita duas chamadas concorrentes a `RegistrarMovimentacao.Execute`. Os testes de concorrência em `infra/movimentacao_repository_test.go:160` e `:233` reimplementam a regra dentro do próprio teste (`:172-184`). Por isso, mover a checagem de saldo para fora da transação (`app/registrar_movimentacao.go:127-137`) passa em toda a suíte, e isso reabre a corrida que AD-009 proíbe: duas saídas da última unidade podem ambas ser aceitas.
   - **Correção sugerida** (não aplicada): um teste de integração com 2 goroutines chamando `Execute` com SAIDA de 1 sobre saldo 1, esperando 1 sucesso, 1 `ErrSaldoInsuficiente` e saldo final 0.
2. **M5 sobreviveu (MOV-03 AC2, EST-D-009).** Nenhum teste tenta uma devolução que referencie outra `DEVOLUCAO`. `TestRegistrarMovimentacaoDevolucaoWithAnInvalidReferenceIsRejected` (`app/registrar_movimentacao_test.go:206`) cobre só inexistente, outro produto e AJUSTE. A regra pode ser afrouxada para aceitar devolução de devolução sem nenhum teste falhar.
3. **Fraco: PRD-01 AC4.** O AC exige "nunca revelando se o `codigo` já existe", mas `app/criar_produto_test.go:69` usa um `codigo` que não existe. Nenhum teste cria o código antes e confirma que o ator sem permissão recebe `forbidden`, e não `codigo_duplicado`. Hoje o comportamento está correto pela ordem do código (`app/criar_produto.go:47`).
4. **Fraco: PRD-02 AC1.** A spec diz "exatamente como criados", mas os testes só conferem a quantidade (`app/listar_produtos_test.go:31`) ou a presença do id (`http/produtos_handler_test.go:74-82`). Nenhum compara `codigo`, `nome` e `unidade_medida` listados com os criados.
5. **Fraco: API-02 AC5.** Via HTTP só a variante "sem `movimentacao_de_id`" é exercitada (`http/movimentacoes_handler_test.go:82`). As variantes inexistente, outro SKU e AJUSTE/DEVOLUCAO ficam só no app.
6. **Lacunas de precisão da spec:**
   - (a) A edge case de `spec/05-api-http.md:93` exige recusar DEVOLUCAO sem referência "na validação de forma do contrato… antes de qualquer chamada ao caso de uso". O design (`design/05-api-http.md:59`, `:123`) decidiu o contrário: o caso de uso valida e devolve `devolucao_invalida`. Esse é o comportamento implementado e testado (`http/movimentacoes_handler_test.go:82`). Falta atualizar a spec.
   - (b) PRD-01 AC3 diz só "um erro de validação". Na prática há dois códigos: `validation_failed` (string vazia, pelo contrato) e `campo_obrigatorio` (só espaços, pelo EST-D-013). Esse segundo não está em EST-D-006, que fixa 7 códigos (`spec/05-api-http.md:85`). O STATE marca EST-D-013 como "sujeita a veto do mantenedor".
   - (c) As tabelas de rastreabilidade de `spec/04-permissoes.md:55` e `spec/05-api-http.md:101-104` ainda dizem "Design | Pending".
7. **Teste com nome enganoso.** `TestGetSaldoWithOnlyMovimentacaoReadPermissionIsForbidden` (`http/ajustes_saldo_handler_test.go:113`) usa ASSOCIADO, que não tem nenhuma permissão de estoque. O teste não prova o que o nome diz; a prova real está em `app/consultar_saldo_test.go:57`. No mesmo arquivo, o teste de produto duplicado tem o comentário "PRD-01 AC3" (`http/produtos_handler_test.go:47`), quando deveria ser AC2.
8. **O teste de concorrência de saldo não discrimina sozinho.** Com M16 aplicado, `TestConcurrentSaidasDisputingTheLastUnitNeverBothSucceed` (`infra/movimentacao_repository_test.go:160`) passou. Só o teste determinístico `:233` matou o mutante. Não é falha, porque o teste determinístico existe, mas o teste de outcome sozinho não protege a regra.
9. **Todos os testes das regras de negócio de `estoque/app` exigem Docker** (`//go:build integration`). `go test ./...` sem a tag mostra `estoque/app [no test files]`, então a regra de saldo negativo, o sinal da devolução e o motivo obrigatório não são verificados no gate rápido.

## Confirmação de isolamento

- **Baseline**, antes de qualquer mutação, com `git status --porcelain` na árvore real:
  ```
   M web/lib/api/financeiro.d.ts
   M web/lib/api/identity.d.ts
   M web/lib/api/platform.d.ts
  ```
- **Mutações:** só na worktree destacada `<scratchpad>/wt-mut`, fora do repositório. Ao final, a worktree mostrava `git status --porcelain` vazio. Ela foi removida com `git worktree remove --force` e `git worktree prune`. `git worktree list` mostra só a árvore principal. Nenhum `git stash` foi usado.
- **Árvore real depois da remoção**, antes de escrever este arquivo: idêntica ao baseline (as mesmas 3 linhas). HEAD continua `17ba7639d55c85d2577e3bff5dbe9533f9c2b8e1`.
- **Única alteração na árvore real:** a criação deste arquivo, `.specs/features/estoque/validation.md`.

---

## Re-verificação 1 (ab7eae9)

**Veredito: PASS.**

### Escopo do diff

`git diff 17ba763 ab7eae9 --stat` mostra 7 arquivos, todos `*_test.go`, com +214/−16:

- `app/criar_produto_test.go`
- `app/listar_produtos_test.go`
- `app/registrar_movimentacao_lock_test.go` (arquivo novo)
- `app/registrar_movimentacao_test.go`
- `http/ajustes_saldo_handler_test.go`
- `http/movimentacoes_handler_test.go`
- `http/produtos_handler_test.go`

O código de produção não mudou. HEAD é `ab7eae991a80c9d07a22a37ecb80694c52f9616b`.

### Mutantes reaplicados

As mutações são as mesmas da primeira passada, aplicadas numa worktree destacada nova (`<scratchpad>/wt-mut2`).

| # | Mutação | Resultado |
| --- | --- | --- |
| M5 | `(ref.Tipo != domain.Entrada && ref.Tipo != domain.Saida)` → `(ref.Tipo == domain.Ajuste)` (`app/registrar_movimentacao.go:114`) | **Morto**: `TestRegistrarMovimentacaoDevolucaoWithAnInvalidReferenceIsRejected` (asserção nova em `app/registrar_movimentacao_test.go:253-267`) e `TestCreateMovimentacaoDevolucaoWithAnInvalidReferenceAnswers422` (`http/movimentacoes_handler_test.go:96`, caso `outra_devolucao` em `:113`) |
| M18 | checagem `SaldoComLock` + `saldo+qtd < 0` movida para fora de `r.Tx` (`app/registrar_movimentacao.go:127-137`) | **Morto**: `TestRegistrarMovimentacaoSaidaLocksTheBalanceInsideTheSameTransactionAsTheInsert` (`app/registrar_movimentacao_lock_test.go:49`, determinístico: `:70-75`) e `TestConcurrentSaidasThroughTheUseCaseNeverBothTakeTheLastUnit` (`:82`) |
| M16 | `pg_advisory_xact_lock(hashtext(?))` → `SELECT 1` (`infra/movimentacao_repository.go:97`) | **Morto**: `TestSaldoComLockSerializesConcurrentCallersOfTheSameProduto` (infra) e `TestConcurrentSaidasThroughTheUseCaseNeverBothTakeTheLastUnit` |

Sobre M16 e o teste novo de concorrência pelo caso de uso:

- **Ele mata o mutante sozinho:** rodado com `-run` isolado, 4 de 4 execuções falharam com `rodada 1: 2 sucessos e 0 saldo_insuficiente` (`app/registrar_movimentacao_lock_test.go:123`).
- **A morte depende de agendamento:** o teste depende de colisão real entre goroutines e repete 10 rodadas. Ele é complementado pelo teste determinístico de infra.

**Placar:** os 18 mutantes da primeira passada estão mortos. As outras 15 mutações não foram reaplicadas porque o código de produção não mudou e nenhum teste foi removido.

### ACs antes fracos ou parciais

| AC | Novo teste (`file:line`) | Valor afirmado | Bate com a spec? |
| --- | --- | --- | --- |
| PRD-01 AC4 | `app/criar_produto_test.go:70` (setup cria `EXISTENTE`, laço em `:78` com `NOVO` e `EXISTENTE`) | `ErrForbidden` para os dois códigos; para o que já existe, nunca `codigo_duplicado` | Sim |
| PRD-02 AC1 | `app/listar_produtos_test.go:44` | os 3 itens listados têm `Codigo`, `Nome`, `UnidadeMedida` e `CriadoEm` iguais aos criados (as unidades são distintas, `UN-A/B/C`) | Sim |
| API-02 AC5 | `http/movimentacoes_handler_test.go:96-121` | `422 devolucao_invalida` para inexistente, outro SKU, AJUSTE e outra DEVOLUCAO | Sim |
| MOV-02 AC5 | `app/registrar_movimentacao_lock_test.go:49` (lock e insert na mesma tx) e `:82` (2 saídas concorrentes via `Execute`, 10 rodadas) | 1 sucesso, 1 `ErrSaldoInsuficiente`, saldo 0; `lockTx == criarTx` | Sim |
| MOV-03 AC2 | `app/registrar_movimentacao_test.go:253-267` | referenciar outra DEVOLUCAO → `ErrDevolucaoInvalida` | Sim |

**Contagem atualizada:** 44 dos 45 ACs têm teste que bate com a spec. O 45º, PERM-01 AC3, é estrutural e foi verificado por inspeção, garantido pelo compilador. Nenhum AC ficou sem teste, fraco ou parcial.

O achado 7 da primeira passada também foi resolvido:

- o teste HTTP foi renomeado para `TestGetSaldoWithoutThePermissionIsForbidden` (`http/ajustes_saldo_handler_test.go:115`), com o motivo explicado no comentário;
- o comentário de `http/produtos_handler_test.go:47` foi corrigido para AC2.

### Item (c): tabelas de rastreabilidade

Estado conferido no working tree com `git diff 17ba763 -- .specs/features/estoque/spec`:

- `spec/04-permissoes.md`: PERM-01 está como `Implementing | Verified`.
- `spec/05-api-http.md`: API-01 a API-04 estão como `Implementing | Verified`.

As alterações estão corretas e coerentes com 01-03. Elas ainda não foram commitadas e o autor é o responsável por elas.

### Gates (árvore real, `ab7eae9`)

| Gate | Resultado |
| --- | --- |
| `go vet ./...` | saída 0 |
| `go test -tags=integration -count=1 -p 1 ./internal/estoque/... ./internal/httpapi/...` | `estoque` ok, `estoque/app` ok (6.1s), `estoque/http` ok (7.2s), `estoque/infra` ok (4.4s), `httpapi` ok (14.4s) |
| Estabilidade dos 2 testes novos de lock (`-count=3`) | ok |
| `golangci-lint run --path-mode=abs --build-tags=integration` | `0 issues.` |

### Itens abertos

**Para o mantenedor** (redação da spec, não bloqueiam o PASS):

- (a) `spec/05-api-http.md:93` contradiz o design aprovado (`design/05-api-http.md:59`, `:123`). O autor deixou para o mantenedor, pela regra de parada em caso de conflito entre documentos aprovados.
- (b) PRD-01 AC3 é vago diante dos dois códigos reais, e EST-D-013 ainda está sujeita a veto.
- O achado 9, regras de `app` cobertas só por testes de integração, é convenção documentada do projeto (STATE.md).

**Novo, menor e não bloqueante:**

- `gofmt -l ./internal/estoque` lista `app/registrar_movimentacao_lock_test.go`. É o alinhamento dos campos de `txRecordingStore` em `:21-28`. O `golangci-lint` não acusa porque o gofmt não está habilitado nele.

### Isolamento

- **Baseline desta rodada:**
  ```
   M .specs/features/estoque/spec/04-permissoes.md
   M .specs/features/estoque/spec/05-api-http.md
   M web/lib/api/financeiro.d.ts
   M web/lib/api/identity.d.ts
   M web/lib/api/platform.d.ts
  ?? .specs/features/estoque/validation.md
  ```
- **Mutações:** só em `<scratchpad>/wt-mut2`, uma worktree destacada em `ab7eae9`. Cada arquivo mutado foi restaurado depois do teste, e a worktree terminou com porcelain vazio antes de ser removida com `git worktree remove --force` e `git worktree prune`. Nenhum `git stash` foi usado.
- **Árvore real depois da rodada:** idêntica ao baseline. A única alteração é a atualização deste arquivo, que já estava como `??`.

## Nota posterior (2026-10-04) — EST-D-013 resolvida pela Opção A

Esta verificação descreve o estado em `ab7eae9`, quando `ErrCampoObrigatorio` respondia `422 campo_obrigatorio` (linha de PRD-01 AC3 e item 6b acima). O mantenedor não aprovou esse code novo: `EST-D-013` foi resolvida reutilizando o erro transversal `422 validation_failed`, sem alterar `EST-D-006`. Depois dessa mudança, `"` e valores só com espaços (incluindo U+00A0) nos 3 campos respondem o mesmo `422 validation_failed` (`api/internal/estoque/http/produtos_handler_test.go:36`), e a tabela `sentinels` tem exatamente os 7 codes de `EST-D-006` (`api/internal/estoque/http/handler_internal_test.go:46`). Mutações do autor sobre a mudança, todas mortas: remover o ramo de `writeError` (voltaria a 500), voltar a `campo_obrigatorio` e tirar `estoqueSpec` do validador de contrato. O veredito PASS se mantém; o item 6b fica resolvido.
