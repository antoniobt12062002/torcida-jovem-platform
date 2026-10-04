# Movimentações de Estoque Design

**Spec**: `.specs/features/estoque/spec/02-movimentacoes.md`
**Status**: Approved

---

## Architecture Overview

Um único caso de uso (`RegistrarMovimentacao`) cobre `ENTRADA`/`SAIDA`/`DEVOLUCAO` (discriminados por `Tipo`), mesma decisão que `financeiro/app.CriarLancamento` toma para `RECEITA`/`DESPESA` — a diferença de sinal e de pré-condição é lógica interna do caso de uso, não três casos de uso separados. `AjustarEstoque` fica em `03-ajustes-e-saldo` (permissão e regra de bloqueio diferentes).

```mermaid
graph TD
    A[RegistrarMovimentacao.Execute] --> B{Tipo}
    B -->|ENTRADA| C[quantidade positiva, sem lock]
    B -->|SAIDA| D[quantidade negativa, com lock+checagem]
    B -->|DEVOLUCAO| E{Movimentação referenciada}
    E -->|SAIDA| C2[quantidade positiva, sem lock]
    E -->|ENTRADA| D2[quantidade negativa, com lock+checagem]
    D --> F[pg_advisory_xact_lock]
    D2 --> F
    F --> G[SUM(quantidade) do produto]
    G --> H{saldo+quantidade >= 0?}
    H -->|sim| I[Criar + Auditar na mesma transação]
    H -->|não| J[ErrSaldoInsuficiente, rollback]
    C --> I
    C2 --> I
```

---

## Code Reuse Analysis

### Existing Components to Leverage

| Component | Location | How to Use |
| --- | --- | --- |
| `pg_advisory_xact_lock` | já usado em `identity/infra/role_repository.go:60`, `user_repository.go:187` | Lock consultivo por `produto_id` (via `hashtext(produto_id::text)`), preso à transação (`_xact_`, libera automaticamente no commit/rollback) — resolve a concorrência que o próprio `ADR-007` deixou como pergunta aberta, sem técnica nova. |
| `platform/database.WithTx` | `api/internal/platform/database/` | Lock+checagem+inserção+auditoria na mesma transação. |
| `estoque/app.ProdutoReader` (porta de `01-produtos`) | `api/internal/estoque/app/criar_produto.go` (ou arquivo de porta compartilhado) | Valida que `produto_id` existe antes de qualquer lock. |
| `platform/audit` | — | `movimentacao.create` reaproveitado por `ENTRADA`/`SAIDA`/`DEVOLUCAO` — mesma decisão de `FIN-D-003` (devolução não é ação de auditoria distinta). |

### Integration Points

| System | Integration Method |
| --- | --- |
| `01-produtos` | Lê via `ProdutoReader.Buscar` — nunca escreve. |
| `03-ajustes-e-saldo` | Lê `movimentacoes_estoque` (mesma tabela) via `SaldoReader`/`MovimentacaoLister` — escreve na mesma tabela com `Tipo=AJUSTE`, mas em caso de uso e permissão próprios. |
| Banco (Postgres) | Nova migration `000009_estoque_movimentacoes` (depende de `000008`). |

---

## Components

### `estoque/domain.Movimentacao`

- **Purpose**: entidade da movimentação — append-only, nunca editada.
- **Location**: `api/internal/estoque/domain/movimentacao.go`
- **Interfaces**:

```go
type TipoMovimentacao string
const (
    Entrada   TipoMovimentacao = "ENTRADA"
    Saida     TipoMovimentacao = "SAIDA"
    Ajuste    TipoMovimentacao = "AJUSTE"
    Devolucao TipoMovimentacao = "DEVOLUCAO"
)

type OrigemMovimentacao string
const (
    OrigemVenda         OrigemMovimentacao = "VENDA"
    OrigemCompra        OrigemMovimentacao = "COMPRA"
    OrigemEvento        OrigemMovimentacao = "EVENTO"
    OrigemInventario    OrigemMovimentacao = "INVENTARIO"
    OrigemAjusteManual  OrigemMovimentacao = "AJUSTE_MANUAL"
)

type Movimentacao struct {
    ID               string
    ProdutoID        string
    Tipo             TipoMovimentacao
    Quantidade       int64  // já assinada: positiva ou negativa, nunca zero
    Origem           OrigemMovimentacao
    Motivo           *string // só preenchido em AJUSTE
    MovimentacaoDeID *string // só preenchido em DEVOLUCAO
    ResponsavelID    string
    CriadoEm         time.Time
}
```

- **Dependencies**: nenhuma.
- **Reuses**: mesmo formato de `financeiro/domain.Lancamento` (campos simples, `*string` para opcionais).

Sentinels:

```go
ErrMovimentacaoNaoEncontrada = errors.New("movimentação não encontrada")
ErrDevolucaoInvalida         = errors.New("devolução inválida")
ErrSaldoInsuficiente         = errors.New("saldo insuficiente")
ErrQuantidadeInvalida        = errors.New("quantidade inválida")
```

### `app.RegistrarMovimentacao`

- **Purpose**: implementa MOV-01/MOV-02/MOV-03 — entrada, saída e devolução, com o sinal e a checagem de saldo corretos por tipo.
- **Location**: `api/internal/estoque/app/registrar_movimentacao.go`
- **Interfaces**:

```go
type RegistrarMovimentacaoInput struct {
    Actor            authz.Principal
    Tipo             domain.TipoMovimentacao // ENTRADA | SAIDA | DEVOLUCAO (nunca AJUSTE aqui)
    ProdutoID        string
    Quantidade       int64 // sempre positivo na entrada do caso de uso
    Origem           domain.OrigemMovimentacao
    MovimentacaoDeID *string // obrigatório só quando Tipo=DEVOLUCAO
}
func (r *RegistrarMovimentacao) Execute(ctx, in RegistrarMovimentacaoInput) (domain.Movimentacao, error)
```

- **Dependencies**: `Authz Authorizer`, `Produtos ProdutoReader`, `Movimentacoes MovimentacaoStore` (porta: `Buscar`, `Criar`, `SaldoComLock`), `Audit Auditor`, `Tx TxFunc`.
- **Reuses**: template de `financeiro/app.CriarLancamento` (permissão → validação de referência → transação com lock condicional → persistir+auditar).

**Lógica do sinal** (EST-D-008/EST-D-009, já fechadas na spec):

```go
if in.Quantidade <= 0 { return ErrQuantidadeInvalida }

var qtd int64
var movDeID *string
switch in.Tipo {
case domain.Entrada:
    qtd = in.Quantidade
case domain.Saida:
    qtd = -in.Quantidade
case domain.Devolucao:
    ref, err := r.Movimentacoes.Buscar(ctx, *in.MovimentacaoDeID)
    if errors.Is(err, domain.ErrMovimentacaoNaoEncontrada) { return domain.ErrDevolucaoInvalida }
    if err != nil { return err }
    if ref.ProdutoID != in.ProdutoID || (ref.Tipo != domain.Entrada && ref.Tipo != domain.Saida) {
        return domain.ErrDevolucaoInvalida
    }
    if ref.Tipo == domain.Saida { qtd = in.Quantidade } else { qtd = -in.Quantidade }
    movDeID = in.MovimentacaoDeID
}
```

**Checagem de saldo** (só quando `qtd < 0`, dentro da mesma transação do `Criar`):

```go
err = r.Tx(ctx, func(ctx context.Context) error {
    if qtd < 0 {
        saldo, err := r.Movimentacoes.SaldoComLock(ctx, in.ProdutoID) // pg_advisory_xact_lock + SUM
        if err != nil { return err }
        if saldo+qtd < 0 { return domain.ErrSaldoInsuficiente }
    }
    criada, err := r.Movimentacoes.Criar(ctx, domain.Movimentacao{
        ProdutoID: in.ProdutoID, Tipo: in.Tipo, Quantidade: qtd, Origem: in.Origem,
        MovimentacaoDeID: movDeID, ResponsavelID: in.Actor.UserID,
    })
    if err != nil { return err }
    return r.Audit.Record(ctx, audit.Entry{
        ActorType: audit.ActorUser, ActorID: in.Actor.UserID, Action: ActionMovimentacaoCreate,
        EntityType: "movimentacao_estoque", EntityID: criada.ID, Outcome: audit.OutcomeSuccess,
        After: map[string]any{"tipo": string(criada.Tipo), "produto_id": criada.ProdutoID, "quantidade": criada.Quantidade},
    })
})
```

```go
const PermMovimentacaoCreate authz.Permission = "estoque:movimentacao:create"
const ActionMovimentacaoCreate audit.Action = "movimentacao.create"
```

### `app.ListarMovimentacoes`

- **Purpose**: implementa MOV-04.
- **Location**: `api/internal/estoque/app/listar_movimentacoes.go`
- **Interfaces**: `ListarMovimentacoesInput{Actor authz.Principal; ProdutoID string}`; `Execute(ctx, in) ([]domain.Movimentacao, error)`.
- **Dependencies**: `Authz Authorizer`, `Movimentacoes MovimentacaoLister` (porta: `ListarPorProduto(ctx, produtoID) ([]domain.Movimentacao, error)`).
- **Reuses**: template de `financeiro/app.ListarLancamentos`.

```go
const PermMovimentacaoRead authz.Permission = "estoque:movimentacao:read"
```

### `infra.MovimentacaoRepository`

- **Purpose**: satisfaz `MovimentacaoStore`/`MovimentacaoLister`/`SaldoReader` (a última consumida também por `03-ajustes-e-saldo`).
- **Location**: `api/internal/estoque/infra/movimentacao_repository.go`
- **Interfaces**: `Criar`, `Buscar`, `ListarPorProduto`, `SaldoComLock(ctx, produtoID) (int64, error)`, `Saldo(ctx, produtoID) (int64, error)` (sem lock, usado por `ConsultarSaldo` em `03`).
- **Dependencies**: `*gorm.DB`.
- **Reuses**: estilo de `financeiro/infra.LancamentoRepository`; `SaldoComLock` reaproveita o padrão de `pg_advisory_xact_lock` já usado em `identity/infra`.

```go
func (r *MovimentacaoRepository) SaldoComLock(ctx context.Context, produtoID string) (int64, error) {
    tx := r.db.WithContext(ctx)
    if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", produtoID).Error; err != nil {
        return 0, err
    }
    var saldo int64
    err := tx.Raw("SELECT COALESCE(SUM(quantidade), 0) FROM movimentacoes_estoque WHERE produto_id = ?::uuid", produtoID).Scan(&saldo).Error
    return saldo, err
}
```

`hashtext(produtoID)` serializa toda escrita negativa do mesmo SKU — duas saídas concorrentes do último item nunca correm em paralelo: a segunda espera a transação da primeira terminar (commit ou rollback), então vê o saldo já atualizado. `FOR UPDATE` não é usado aqui porque o saldo é uma agregação (`SUM`), e Postgres não permite `FOR UPDATE` sobre uma consulta com função de agregação — `pg_advisory_xact_lock` é a ferramenta correta para serializar por chave lógica (`produto_id`), já precedente no próprio código.

---

## Data Models

### Migration `000009_estoque_movimentacoes`

```sql
CREATE TABLE movimentacoes_estoque (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    produto_id uuid NOT NULL REFERENCES produtos_estoque(id),
    tipo text NOT NULL CHECK (tipo IN ('ENTRADA', 'SAIDA', 'AJUSTE', 'DEVOLUCAO')),
    quantidade bigint NOT NULL CHECK (quantidade <> 0),
    origem text NOT NULL CHECK (origem IN ('VENDA', 'COMPRA', 'EVENTO', 'INVENTARIO', 'AJUSTE_MANUAL')),
    motivo text,
    movimentacao_de_id uuid REFERENCES movimentacoes_estoque(id),
    responsavel_id uuid NOT NULL REFERENCES users(id),
    criado_em timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_movimentacoes_estoque_produto_id ON movimentacoes_estoque(produto_id);
```

Sem `atualizado_em` — a tabela é append-only por desenho (`AD-009`); nenhuma linha é alterada depois de criada, então não há o que marcar como atualizado. Índice em `produto_id`: toda consulta de saldo e toda listagem filtram por ele.

---

## Error Handling Strategy

| Error Scenario | Handling | User Impact |
| --- | --- | --- |
| Saldo insuficiente em `SAIDA`/`DEVOLUCAO`-de-`ENTRADA` | `domain.ErrSaldoInsuficiente`, dentro da transação, antes do `Criar` | `05-api-http` mapeia para `409 saldo_insuficiente` |
| `movimentacao_de_id` inválido para `DEVOLUCAO` | `domain.ErrDevolucaoInvalida` | `422 devolucao_invalida` |
| `produto_id` inexistente | `domain.ErrProdutoNaoEncontrado` (de `01-produtos`) | `404 produto_nao_encontrado` |
| `quantidade <= 0` | `domain.ErrQuantidadeInvalida` | `422 quantidade_invalida` |

---

## Risks & Concerns

| Concern | Location | Impact | Mitigation |
| --- | --- | --- | --- |
| Concorrência real de saldo (duas saídas disputando a última unidade) | `infra/movimentacao_repository.go:SaldoComLock` (a criar) | Alto se não testado — é exatamente a pergunta que `ADR-007` deixou aberta | Task dedicada de teste de integração real (2 goroutines, Postgres real via `testutil.NewTestDB`, confirmando que só uma das duas conclui com sucesso) — exigido explicitamente no "Done when" da task correspondente. |
| `hashtext()` pode colidir entre dois `produto_id` diferentes (hash de 32 bits) | mesmo ponto acima | Baixo — uma colisão só serializaria duas operações de SKUs diferentes sem necessidade (nunca causa inconsistência, só uma espera desnecessária rara); `identity/infra` já aceita esse mesmo trade-off para seus locks. | Nenhuma mitigação adicional — consistente com o padrão já aceito no código. |

---

## Tech Decisions

| Decision | Choice | Rationale |
| --- | --- | --- |
| Lock de concorrência | `pg_advisory_xact_lock(hashtext(produto_id))`, nunca `SELECT ... FOR UPDATE` | Postgres não permite `FOR UPDATE` com `SUM()`; o advisory lock por chave lógica já é padrão aceito no código (`identity/infra`). |
| Devolução e entrada/saída compartilham um único caso de uso | `RegistrarMovimentacao` com `Tipo` | Decisão do mantenedor (M.5/EST-D-005): um único endpoint HTTP, logo um único caso de uso por trás — evita duplicar a lógica de lock+checagem em dois lugares. |

---

## Tips

A checagem de saldo e a inserção **sempre** ocorrem na mesma transação do lock — nunca o lock numa chamada e a checagem/inserção noutra. Qualquer refatoração futura deve preservar isso.
