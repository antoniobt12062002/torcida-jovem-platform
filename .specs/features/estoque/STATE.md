# Estoque — STATE

Documento de coordenação da feature `estoque`, decomposta em 5 sub-specs (`01`-`03` o domínio V1, `04` RBAC, `05` sua camada HTTP) — mesmo padrão estrutural de `financeiro/STATE.md`. Mantém enxuto o `.specs/STATE.md` do projeto: aqui ficam as decisões e o mapa específicos desta feature.

Convenção de numeração local: **EST-D-NNN** para decisões que valem só para `estoque`. Nunca reaproveita o numerador `AD-NNN` do `.specs/STATE.md` global.

---

## Mapa das sub-specs

| # | Spec | Requisitos | Status |
|---|---|---|---|
| 01 | `spec/01-produtos.md` | PRD-01, PRD-02 | Concluída (T1-T5) |
| 02 | `spec/02-movimentacoes.md` | MOV-01 a MOV-04 | Concluída (T1-T5) |
| 03 | `spec/03-ajustes-e-saldo.md` | AJS-01, AJS-02 | Concluída (T1-T3) |
| 04 | `spec/04-permissoes.md` | PERM-01 | Concluída (T1-T2) |
| 05 | `spec/05-api-http.md` | API-01 a API-04 | Especificada; não implementada |

## Dependências

```
01-produtos
       │  (produto_id)
       ▼
02-movimentacoes ──► 03-ajustes-e-saldo   (mesma tabela movimentacoes_estoque; permissão e regra de bloqueio diferentes)

04-permissoes: transversal — depende de 01-03 publicarem suas constantes Perm*, não depende de nenhuma delas em código real.
05-api-http: depende de 01-04 completas (precisa dos 6 casos de uso e das 6 permissões).
```

Nenhuma dependência circular. `02⇢03` é unidirecional de verdade (03 lê a mesma tabela e o mesmo repositório de 02, nunca o contrário) — mais simples que a relação `02⇢01` de `financeiro` (que tinha uma leitura pontual em sentido contrário à ordem de implementação); aqui a ordem de dependência e a ordem de implementação coincidem.

## Fronteiras de responsabilidade

| Dado/operação | Dona | Nota |
|---|---|---|
| `Produto` (SKU mínimo de estoque) | `01` | única spec que insere; nunca editado nem desativado (`EST-D-007`) |
| `Movimentacao` (todos os tipos, incluindo `AJUSTE`) | `02` | única spec que insere a linha (inclusive quando `03` cria um `AJUSTE` — usa o mesmo repositório de `02`, com uma porta mais estreita) |
| Regra de bloqueio de saldo negativo (`SaldoComLock`) | `02` | a regra de quando bloquear (saída e devolução-de-entrada) é de `02`; `03` explicitamente não a usa |
| Catálogo de permissões e matriz de papéis | `04` | única spec que define a `Contribution` |

`01`-`03` são sub-specs do mesmo módulo Go (`api/internal/estoque/`), não módulos de negócio separados — mesma disciplina de `financeiro`.

### Estrutura de pacotes

Mesmo esqueleto de `architecture-overview.md`, idêntico ao de `identity`/`financeiro`:

```
api/internal/estoque/
├── module.go     # estoque.Contribution() (04) e estoque.New(...) (05)
├── domain/       # produto.go, movimentacao.go
├── app/          # um arquivo por caso de uso
├── infra/        # produto_repository.go, movimentacao_repository.go
└── http/         # vazio até 05-api-http
```

---

## Decisões fechadas (EST-D-NNN)

### EST-D-001 — SKU mínimo é dono por `estoque`; catálogo comercial fica para `loja` — **Aprovada**
**Decisão**: `estoque` possui uma representação mínima de SKU (código, nome, unidade de medida) — nenhum dado comercial (preço, descrição, imagem, categoria). O catálogo comercial, quando `loja` for especificada, poderá referenciar o SKU criado aqui por id. Nenhum módulo `produtos` independente é criado agora.
**Escopo**: `01-produtos`.

### EST-D-002 — Sem reserva no V1 — **Aprovada**
**Decisão**: nenhum mecanismo de reserva, expiração de reserva ou carrinho reservado. A responsabilidade por decidir quando uma venda deve reservar estoque fica para a futura especificação de `loja`. `estoque` V1 só tem saldo e movimentações.
**Escopo**: `02-movimentacoes` (não implementa nada disso).

### EST-D-003 — Controle por SKU atômico, sem variação estruturada — **Aprovada**
**Decisão**: cada SKU representa uma unidade distinta de estoque. Tamanho, cor ou outra característica que exija saldo independente resulta em um SKU diferente (ex. `CAMISA-TJ-P-PRETA`, `CAMISA-TJ-M-PRETA` são SKUs independentes). Sem `produto_variacao`, sem relação pai/filho, sem tabela de atributos de variação.
**Escopo**: `01-produtos`.

### EST-D-004 — RBAC: 6 permissões — **Aprovada**
**Decisão**: `estoque:produto:create`, `estoque:produto:read`, `estoque:movimentacao:create`, `estoque:movimentacao:adjust`, `estoque:movimentacao:read`, `estoque:saldo:read`. `ESTOQUE_LOJA` recebe as 6; `DIRETORIA` e `CONSELHO_FISCAL` recebem as 3 de leitura; `PRESIDENTE` automático; demais papéis nenhuma. A separação de `movimentacao:create`/`movimentacao:adjust` é intencional — ajuste é operação excepcional, com motivo obrigatório e autorização própria.
**Escopo**: `04-permissoes`.

### EST-D-005 — API de movimentações: endpoint único com `tipo` + endpoint próprio de ajuste — **Aprovada**
**Decisão**: `POST /api/v1/estoque/movimentacoes` com `tipo` (`ENTRADA`\|`SAIDA`\|`DEVOLUCAO`) no corpo; `POST /api/v1/estoque/ajustes` como endpoint próprio. Sem `/entradas`/`/saidas` separados. Devolução referencia a movimentação original por `movimentacao_de_id`; a original nunca é editada.
**Escopo**: `05-api-http`.

### EST-D-006 — Mapeamento de erros HTTP — **Aprovada**
**Decisão**: `saldo_insuficiente`→`409` (conflito com o estado atual do estoque, não erro estrutural de payload); `motivo_obrigatorio`→`422`; `produto_nao_encontrado`→`404`; `forbidden`→`403`. Fechados no design, consistentes com a mesma régua semântica já usada em `financeiro` (`FIN-D-022`: recurso inexistente→404, conflito com estado atual→409, validação de campo/referência→422): `codigo_duplicado`→`409` (conflito de identidade, análogo a `financeiro:conta_ja_utilizada`), `devolucao_invalida`→`422` (validação de referência de campo), `quantidade_invalida`→`422`.
**Escopo**: `05-api-http` (API-04).

### EST-D-007 — SKU imutável após a criação — **Aprovada**
**Decisão**: sem `produto:update`, sem `produto:deactivate`, sem exclusão física. O campo `ativo` é removido do modelo (não há funcionalidade que o justifique no V1). A identidade do SKU é estável após a criação. Desativação/aposentadoria de SKU, se necessária no futuro, é evolução própria, com nova decisão.
**Escopo**: `01-produtos`.

### EST-D-008 — Sinal da devolução inverte o sentido da movimentação referenciada (achado lógico, registrado com justificativa)
**Achado**: nenhum documento (`ADR-007` incluído) define a direção do efeito de uma devolução. **Decisão derivada, não uma regra de negócio nova**: devolver uma `SAIDA` repõe estoque (soma positiva, nunca bloqueada); devolver uma `ENTRADA` remove estoque (soma negativa, sujeita à mesma regra de bloqueio de saldo negativo que `SAIDA`). Mantém uma única regra de bloqueio (bloqueia quando o efeito líquido é retirar, nunca quando é repor), sem introduzir o conceito de "devolução que nunca falha" nem duplicar a lógica de bloqueio por tipo.
**Escopo**: `02-movimentacoes`.

### EST-D-009 — Devolução só pode referenciar `ENTRADA` ou `SAIDA` (achado lógico, registrado com justificativa)
**Achado**: evita cadeias de devolução-de-devolução e evita devolver um ajuste (que já é a própria ferramenta de correção). Mesmo espírito de `FIN-D-003` em `financeiro` (que restringe o que uma devolução pode referenciar).
**Escopo**: `02-movimentacoes`.

### EST-D-010 — Ajuste é um delta assinado, nunca um valor absoluto (achado lógico, registrado com justificativa)
**Achado**: mantém a filosofia de razão de movimentações (`ADR-007`) — todo registro é um delta, nunca um estado absoluto. Calcular o delta necessário para atingir um valor-alvo seria uma função adicional não solicitada por nenhum documento.
**Escopo**: `03-ajustes-e-saldo`.

### EST-D-011 — Nomes de tabela `produtos_estoque`/`movimentacoes_estoque` (achado mecânico)
**Achado**: evita colisão de nome com uma futura tabela de catálogo comercial em `loja`, que pode usar "produto" num sentido mais rico e incompatível com o SKU mínimo de `estoque` (`EST-D-001`).
**Escopo**: `01-produtos`, `02-movimentacoes`.

---

## Regras herdadas de `AD-009`/`ADR-007` (não redecididas aqui, só referenciadas)

Saldo é sempre derivado das movimentações, nunca uma coluna mutável; movimentações são append-only; nenhuma é editada para corrigir histórico; correção ocorre por nova movimentação de ajuste; saída não pode resultar em saldo negativo; ajuste autorizado pode corrigir o saldo inclusive para negativo; ajuste exige motivo; devolução é uma nova movimentação que referencia a original sem jamais alterá-la. Concorrência tratada por `pg_advisory_xact_lock`, dentro da mesma transação, com teste de integração real obrigatório para duas operações concorrentes disputando o último saldo (exigido explicitamente em `tasks/02-movimentacoes.md`/T3).

## Fora do V1 (não implementar sem nova decisão)

Reserva de estoque (`EST-D-002`); integração real com `financeiro` (compra→entrada) ou `loja` (venda→saída) — nenhum dos dois módulos do outro lado existe; custo de aquisição/métodos de custeio; estoque mínimo/alertas; múltiplos locais/depósitos; quantidade fracionária; variação estruturada (`EST-D-003`); qualquer catálogo comercial (`EST-D-001`).

---

## Matriz de permissões do V1 (consolidada)

| Permissão | `ESTOQUE_LOJA` | `DIRETORIA` | `CONSELHO_FISCAL` | `PRESIDENTE` |
|---|---|---|---|---|
| `estoque:produto:create` | ✓ | | | automático |
| `estoque:produto:read` | ✓ | ✓ | ✓ | automático |
| `estoque:movimentacao:create` | ✓ | | | automático |
| `estoque:movimentacao:adjust` | ✓ | | | automático |
| `estoque:movimentacao:read` | ✓ | ✓ | ✓ | automático |
| `estoque:saldo:read` | ✓ | ✓ | ✓ | automático |

`ASSOCIADO`, `TESOURARIA`, `EVENTOS`, `ADMIN_SISTEMA`: nenhuma permissão de `estoque` no V1.

---

## Ordem real de execução entre sub-specs

A ordem de especificação (`01→02→03→04→05`) coincide com a ordem real de implementação — mais simples que `financeiro`, que teve uma exceção documentada (`01`/T4 dependendo de `02`). Aqui:

1. `01` completa (migration+domínio+infra+2 casos de uso)
2. `02` completa (depende de `01` só pela FK de `produto_id`)
3. `03` completa (depende de `02` pela mesma tabela/repositório)
4. `04` completa (depende de `01`-`03` só para referenciar as constantes `Perm*` já publicadas)
5. `05` completa (depende de `01`-`04`)

Nenhuma PR precisa misturar sub-specs — cada uma é uma unidade de execução própria, mesma regra de `financeiro`.

## Decisões futuras (V2, fora do V1 — não implementar sem nova aprovação)

Ver seção "Fora do V1" acima — repetida aqui por paralelismo com `financeiro/STATE.md`: reserva, integração real com `financeiro`/`loja`, custo de aquisição, estoque mínimo/alertas, múltiplos locais, quantidade fracionária, variação estruturada, catálogo comercial.

## Progresso de implementação

- **`01-produtos`/`T1`**: migration `000008_estoque_produtos` (tabela `produtos_estoque`, `codigo UNIQUE NOT NULL`, `GRANT INSERT, SELECT` — nunca `UPDATE`/`DELETE`, `EST-D-007`). Teste de migração dedicado (`estoque_produtos_migration_test.go`, 5 testes: coluna armazenada, unicidade de código, grants corretos, down/up roundtrip, falha clara sem `tj_app`). Nenhuma alteração em outras migrations — `produtos_estoque` não é referenciada por nenhuma FK ainda. Gate completo (full, integration, lint, arquitetura) 100% verde, sem exceção.
- **`01-produtos`/`T2`**: `estoque/domain/produto.go` — `Produto{ID, Codigo, Nome, UnidadeMedida, CriadoEm}`, `ErrProdutoNaoEncontrado`, `ErrCodigoDuplicado`. Sem `Ativo`/`AtualizadoEm` (`EST-D-007`). Primeiro arquivo de `api/internal/estoque/` — `TestProductionCodeRespectsTheModuleBoundaries` já cobre o módulo automaticamente, sem precisar de nenhuma lista a editar.
- **Achado mecânico (`01-produtos`/`T4`)**: a `Test Coverage Matrix` das 5 tasks rotulava os testes de `app` como "unit (com doubles)" — mas o padrão real já estabelecido em `financeiro/app` (`env_test.go`) usa Postgres real (`testutil.NewTestDB`) com `infra` real, só substituindo `Authz` por `simpleAuthz` (decide pela permissão do `Principal`, sem catálogo/RBAC completo) — nunca doubles sem I/O. Corrigido nos 5 arquivos de `tasks/` (rótulo "integration", nunca "unit") para refletir o que de fato é implementado; nenhuma mudança de comportamento, só precisão documental.
- **`01-produtos`/`T3`**: `estoque/infra/produto_repository.go` — `Buscar`/`Criar`/`Listar`, mesmo estilo de `financeiro/infra.ContaRepository` (SQL direto via GORM). Tradução de `codigo` duplicado: nome real da constraint confirmado empiricamente contra o banco (`produtos_estoque_codigo_key`, nomeação implícita do Postgres para `UNIQUE` inline — nunca presumido), usando `errors.AsType[*pgconn.PgError]`, mesmo padrão de `identity/infra.UserRepository` para e-mail duplicado. 4 testes de integração, incluindo o obrigatório de concorrência (2 goroutines criando o mesmo `codigo`: exatamente 1 sucesso e 1 `ErrCodigoDuplicado`, nunca outro resultado) + 1 mutação (nome de constraint trocado) — morta.
- **`01-produtos`/`T4`**: `estoque/app/criar_produto.go` — permissão, validação dos 3 campos (nenhum vazio, `domain.ErrCampoObrigatorio` — sentinel adicionado a `domain/produto.go`, não previsto explicitamente no design, achado mecânico ao implementar PRD-01 AC3), persistência+auditoria `produto.create` na mesma transação. `estoque/app/env_test.go` (novo, mesmo padrão de `financeiro/app/env_test.go`: `simpleAuthz`, `actor()` com UUID sintético — `produtos_estoque` não tem FK para `users`, mesma razão de `financeiro`'s `01-plano-de-contas`). 4 testes de integração (os 4 ACs de PRD-01) + mutação em 2 pontos (validação de campo vazio removida, checagem de permissão removida) — ambos mortos.
- **`01-produtos`/`T5`**: `estoque/app/listar_produtos.go` — consulta pura, sem `Tx`/`Audit`, mesmo template de `financeiro/app.ListarContas`. Sem paginação (mesma decisão deliberada de `FIN-D-021`). 2 testes de integração (os 2 ACs de PRD-02) + mutação (checagem de permissão removida) — morta. **`01-produtos` concluída (T1-T5)** — gate completo (full, integration, lint, arquitetura) 100% verde, sem nenhuma exceção, confirmado nesta tarefa.

- **`02-movimentacoes`/`T1`**: migration `000009_estoque_movimentacoes` (tabela `movimentacoes_estoque`, `tipo`/`origem` via `CHECK`, `quantidade bigint CHECK (quantidade <> 0)`, FKs para `produtos_estoque`, `users` e auto-referência `movimentacao_de_id`; `GRANT INSERT, SELECT` — nunca `UPDATE`/`DELETE`, append-only). Testes de down-migration de `01-produtos` (`estoque_produtos_migration_test.go`) e de `identity` (`identity_migration_test.go`) ajustados para desfazer `000009` antes de `000008`/`000003` respectivamente, mesmo cuidado que `financeiro/01/T1` teve ao introduzir `000007`. 7 testes de migração dedicados (`estoque_movimentacoes_migration_test.go`: coluna armazenada, `CHECK`s, FKs de `produto_id`/`responsavel_id`, auto-referência de devolução, grants corretos, down/up roundtrip, falha clara sem `tj_app`). Gate completo 100% verde (uma falha transiente de Docker-on-Windows em `identity/http`, já documentada como ambiental ao longo de todo este projeto, confirmada passageira ao reexecutar o pacote isoladamente).

- **`02-movimentacoes`/`T2`**: `estoque/domain/movimentacao.go` — `TipoMovimentacao`/`OrigemMovimentacao` (enums), `Movimentacao{ID, ProdutoID, Tipo, Quantidade, Origem, Motivo, MovimentacaoDeID, ResponsavelID, CriadoEm}`, 4 sentinels (`ErrMovimentacaoNaoEncontrada`, `ErrDevolucaoInvalida`, `ErrSaldoInsuficiente`, `ErrQuantidadeInvalida`).

- **`02-movimentacoes`/`T3`**: `estoque/infra/movimentacao_repository.go` — `Buscar`/`Criar`/`ListarPorProduto`/`SaldoComLock` (`pg_advisory_xact_lock(hashtext(produto_id))` + `SUM`, na mesma transação do chamador). **Achado mecânico real**: o teste de outcome pedido literalmente pelo mantenedor (saldo=1, 2 goroutines disputando a última unidade) teve seu mutante (remoção completa do lock) **sobreviver** em 5 execuções consecutivas — a janela de corrida era estreita demais para colidir de forma confiável só por sorte de agendamento do SO. Fortalecido com `TestSaldoComLockSerializesConcurrentCallersOfTheSameProduto`, uma prova determinística e independente de timing de SO: uma segunda chamada ao lock do mesmo `produto_id` fica provadamente bloqueada (medida por tempo) enquanto a primeira mantém a transação aberta, e só desbloqueia depois do commit da primeira — esse teste matou o mutante de forma consistente. Um segundo mutante (trocar `hashtext(produto_id)` por uma chave fixa, tornando o lock global) sobreviveu a todos os testes anteriores — achado real de cobertura, não apenas de timing: nada provava que o lock era por SKU, não global (uma regressão de desempenho séria, serializaria toda escrita de estoque da plataforma). Fortalecido com `TestSaldoComLockDoesNotSerializeCallersOfDifferentProdutos`, que mata esse mutante. 12 testes de integração no total (incluindo os 2 testes de outcome e os 2 de mecanismo) — ambos os mutantes mortos depois das correções.

- **`02-movimentacoes`/`T4`**: `estoque/app/registrar_movimentacao.go` — `RegistrarMovimentacao` cobre `ENTRADA`/`SAIDA`/`DEVOLUCAO` (um único caso de uso, mesma decisão de granularidade de `financeiro/app.CriarLancamento` para `RECEITA`/`DESPESA`, `EST-D-005`). `estoque/app/env_test.go` estendido com `movimentacoes`, `newUser`/`userActor` (`responsavel_id` é FK real para `users`, diferente de `produtos_estoque`) e `criarProdutoDeTeste`/`registrarMovimentacao`. 16 testes de integração (os 4+5+4 ACs combinados de `MOV-01`/`MOV-02`/`MOV-03`) + mutação em 3 pontos (sinal de `SAIDA` invertido, checagem de saldo removida, validação de referência de devolução removida) — a checagem de saldo pegou em tempo de compilação (`saldo` não usado), os outros dois em runtime; todos mortos.

- **`02-movimentacoes`/`T5`**: `estoque/app/listar_movimentacoes.go` — consulta pura, sem paginação, lista qualquer tipo (incluindo `AJUSTE`, criado só em `03`). 2 testes de integração (os 2 ACs de `MOV-04`) + mutação (checagem de permissão removida) — morta. **`02-movimentacoes` concluída (T1-T5)** — gate completo (full, integration, lint, arquitetura) 100% verde, sem nenhuma exceção.

- **`03-ajustes-e-saldo`/`T1`**: `estoque/app/ajustar_estoque.go` — `AjustarEstoque`, permissão própria (`estoque:movimentacao:adjust`, distinta de `create`), ordem permissão→quantidade≠0→motivo→busca→transação (mesmo molde de `financeiro/app.CancelarLancamento`), nunca checa saldo (porta `MovimentacaoCreatorOnly` sem `SaldoComLock` — garantia estrutural). `domain.ErrMotivoObrigatorio` adicionado a `movimentacao.go`. 5 testes de integração (os 5 ACs de AJS-01) + mutação em 2 pontos (motivo removido, permissão removida) — ambos mortos.

- **`03-ajustes-e-saldo`/`T2`**: `estoque/infra/movimentacao_repository.go` estendido com `Saldo` (sem lock) — mesma agregação de `SaldoComLock`, sem a linha do lock, reaproveitado por `ConsultarSaldo` (consulta pura, nunca paga o custo do lock). 2 testes de integração (saldo zero sem movimentação, soma combinada dos 4 tipos) + mutação (`Saldo` fixado em `0`) — morta.

- **`03-ajustes-e-saldo`/`T3`**: `estoque/app/consultar_saldo.go` — `ConsultarSaldo`, consulta pura, permissão própria (`estoque:saldo:read`, distinta de `movimentacao:read`, mesma lógica de `FIN-D-013`). 3 testes de integração (os 3 ACs de AJS-02) + mutação em 2 pontos (permissão removida, checagem de produto removida) — ambos mortos. **`03-ajustes-e-saldo` concluída (T1-T3)** — gate completo (full, integration, lint, arquitetura) 100% verde, sem nenhuma exceção. Domínio V1 de `estoque` (`01`-`03`) está completo — restam só `04-permissoes` e `05-api-http`.

- **`04-permissoes`/`T1`**: `estoque/module.go` — `Contribution()`, modelada direto em `financeiro.Contribution()` (sem bloco institucional, que `estoque` não tem): as 6 permissões de `EST-D-004` como `[]authz.Definition` referenciando as constantes `Perm*` de `estoque/app` (nunca string literal, mesma razão de `FIN-D-019`); `Grants`: `ESTOQUE_LOJA`→as 6, `DIRETORIA`/`CONSELHO_FISCAL`→as 3 de leitura. Nenhuma das 6 é `CommonRead` (mesma decisão de `financeiro`: estoque afeta o razão de movimentações, sempre auditado). `estoque/module_test.go` (novo, mesmo padrão de `financeiro/module_test.go`): 9 testes — contagem exata das 6 sem duplicata, nenhuma `CommonRead`, grant exato por papel (`ESTOQUE_LOJA`/`DIRETORIA`/`CONSELHO_FISCAL`, via `slices.Equal` ordenado), prova própria (belt-and-suspenders) de que `CONSELHO_FISCAL` nunca recebe `create`/`adjust`, nenhum outro papel recebe nada, `BuildMatrix` aceita a `Contribution` isolada, `PRESIDENTE` recebe as 6 automaticamente. 3 mutações (troca de papéis `ESTOQUE_LOJA`↔`DIRETORIA` no `Grants`; remoção de `PermMovimentacaoAdjust` da lista operacional; concessão de `adjust` a `CONSELHO_FISCAL`) — as 3 mortas; a terceira confirmou que o teste belt-and-suspenders pega o problema de forma independente do teste de igualdade exata (os dois falharam juntos, nenhuma dependência oculta entre eles). Gate completo (full, integration, lint, arquitetura) 100% verde, sem nenhuma exceção.
  **Achado (registrado, não corrigido nesta task — fora do escopo de arquivo autorizado)**: `identity/app`'s `forbiddenToConselhoFiscal` (a lista de defesa em profundidade que `BuildMatrix` consulta para recusar qualquer `Contribution` que tente conceder a `CONSELHO_FISCAL` uma ação perigosa) cobre `create`/`update`/`delete`/`cancel`/`deactivate`/`receive`/`pay` — mas não `adjust`, o verbo novo que `estoque` introduz. A `Contribution()` de `estoque` já é segura por construção (nunca concede `adjust` a `CONSELHO_FISCAL`, confirmado pelo teste belt-and-suspenders acima) — este não é um bug ativo. Mas a lacuna estrutural existe: um `Contribution()` futuro (de `estoque` V2 ou de outro módulo) que concedesse `adjust` a `CONSELHO_FISCAL` não seria recusado por `BuildMatrix`, diferente do que já acontece para os outros 7 verbos perigosos. Fechar essa lacuna (acrescentar `"adjust"` a `forbiddenToConselhoFiscal`) seria o mesmo tipo de reforço que `deactivate`/`receive`/`pay` já receberam quando `financeiro` introduziu esses verbos — mas exige editar `api/internal/identity/app/roles_matrix.go`, fora do escopo de arquivos autorizado para esta rodada de `estoque` (`identity/` não está na lista). Não corrigido; registrado aqui para decisão do mantenedor.

- **`04-permissoes`/`T2`**: `estoque.Contribution()` agregado em `cmd/api/main.go` e `cmd/bootstrap-admin/main.go`, junto de `FoundationContributions()`+`financeiro.Contribution()` (mesma linha, mesmo padrão de `FIN-D-020`). `bootstrap_integration_test.go` estendido com a contagem de `permissions WHERE name LIKE 'estoque:%'` (= 6), mesmo padrão do teste equivalente de `financeiro`. Mutação (remoção da linha de wiring em `bootstrap-admin/main.go`) morta em tempo de compilação (import não utilizado) — mesma classe de kill já aceita nesta sessão.
  **Achado real, exatamente da classe prevista por `FIN-D-020`** (consequência técnica direta, corrigida nesta própria task, arquivos fora do escopo original de `estoque` mas dentro da regra de "consequência necessária, sem expansão de escopo"): a suíte de integração completa (`go test -tags=integration ./...`) revelou 5 falhas reais em `identity/infra/role_repository_test.go` só depois do wiring em produção:
  1. O helper `matrix(t, extra...)` (usado por quase todo o arquivo) não agregava `estoque.Contribution()`, divergindo da composição real de produção — corrigido agregando-o, igual a `financeiro.Contribution()`.
  2. Três fixtures sintéticas pré-existentes (`removableFixture()` e duas variações inline), criadas **antes** de `estoque` existir como módulo real, usavam deliberadamente `Module: "estoque"` como "módulo descartável, sem relação com nenhum módulo real" (comentário original). Com `estoque` agora real, esse nome colidia semanticamente (não literalmente — os nomes de permissão `estoque:item:*` nunca colidiam com os 6 reais `estoque:produto:*`/`estoque:movimentacao:*`/`estoque:saldo:*`) com o módulo de produção. Renomeado para `fixture_removivel` (sem hífen: o validador de formato de permissão exige `^[a-z_]+:[a-z_]+:[a-z_]+$`, só letras minúsculas e `_`) nos 3 pontos e nas 4 queries SQL que referenciavam os nomes antigos.
  3. `TestRemovedGrantIsListedAndDeleted` comparava uma sincronização completa (`matrix(t)`, agora com `estoque`) contra uma reduzida reconstruída manualmente que **esqueceu** `estoque.Contribution()` — corrigido agregando-a também, isolando a mudança pretendida pelo teste (remoção do grant de `DIRETORIA` em `identity`) de uma remoção não intencional de todo o módulo `estoque`.
  4. `TestEffectivePermissionsAreTheUnionOfTheActiveRolesPermissions` tinha uma lista `want` fixa da união de permissões de `CONSELHO_FISCAL`+`DIRETORIA` — desatualizada após as duas receberem as 3 leituras reais de `estoque`. Corrigido acrescentando `estoque:movimentacao:read`, `estoque:produto:read`, `estoque:saldo:read` à lista esperada, na mesma ordem alfabética que `slices.Equal` exige.
  Nenhuma mudança de comportamento de produção — só `identity/infra/role_repository_test.go` (arquivo de teste) foi tocado fora do escopo original de arquivos de `estoque`; confirmado por `git diff` que nenhum arquivo de produção de `identity/` mudou. Gate completo (full, integration, lint, arquitetura) 100% verde depois das correções, sem nenhuma exceção. **`04-permissoes` concluída (T1-T2)**.

## Handoff

- **Fase**: Implementação autônoma de `estoque` V1 em andamento sob autorização permanente do mantenedor (2026-10-03). `01-produtos` (T1-T5), `02-movimentacoes` (T1-T5), `03-ajustes-e-saldo` (T1-T3) e `04-permissoes` (T1-T2) concluídas — RBAC de `estoque` publicada e sincronizada nos dois composition roots. Próximo: `05-api-http` (5 tasks). Branch `feature/estoque-v1`, um commit atômico por task, nenhum push/PR/merge realizado ainda.
- **Achado pendente de decisão do mantenedor**: lacuna em `identity/app/roles_matrix.go`'s `forbiddenToConselhoFiscal` (não cobre o verbo `adjust`) — ver nota completa no bullet de `04-permissoes`/T1 acima. Não bloqueia a implementação de `estoque` (a `Contribution()` já é segura por construção), mas é uma alteração fora do escopo de arquivos autorizado nesta rodada.
- **Próximo passo**: `04-permissoes`/T2, depois `05-api-http` (T1-T5), seguindo a DAG e o protocolo por task já em uso (tests-first, mutação nas regras críticas, gate completo, um commit por task), sem pausar para autorização entre tasks, até concluir as 20 tasks de `estoque` V1.
