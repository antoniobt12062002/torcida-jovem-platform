import { MoneyValue } from "@/components/app/money-value";
import { RoleBadge } from "@/components/app/role-badge";
import { activeStatusDef, StatusBadge } from "@/components/app/status-badge";
import { Badge } from "@/components/ui/badge";
import { Card, CardHeader } from "@/components/ui/card";

import { PAPEIS_EXEMPLO, SALDO_NEGATIVO, STATUS_LANCAMENTO, type StatusLancamento } from "./data";
import { SpecimenBlock, SpecimenCaption, SpecimenSection } from "./section";

// Status e valores (DSC-05, DSC-06): cada status com tom, ícone e rótulo,
// nunca só cor; MoneyValue nas quatro variantes; RoleBadge; e o saldo em
// Barlow Condensed e em Geist lado a lado, para a decisão de DD-4.

const ROW = "flex flex-wrap items-center gap-3";

const STATUS_ORDER: readonly StatusLancamento[] = ["CRIADA", "RECEBIDA", "PAGA", "CANCELADA"];

// Valores para comparar a largura dos algarismos (1111 contra 0000 e 8888).
const ALIGNMENT = [111111, 0, 888888] as const;

function DisplayComparison({ family }: { family: "barlow" | "geist" }) {
  // A variante Geist é o mesmo componente, com a família trocada por className (DD-4: Geist 600).
  const className = family === "geist" ? "font-sans font-semibold" : undefined;
  return (
    <Card>
      <CardHeader>
        {/* Parágrafo, não `h4`: o cartão é parte do bloco de título h3. */}
        <p className="text-base leading-6 font-semibold">
          {family === "barlow" ? "Barlow Condensed 700 (proposta)" : "Geist 600 (reserva)"}
        </p>
        <SpecimenCaption>
          {family === "barlow" ? 'MoneyValue size="display"' : 'MoneyValue size="display" className="font-sans font-semibold"'}
        </SpecimenCaption>
      </CardHeader>
      <div className="flex flex-col gap-2">
        <MoneyValue cents={1847390} variant="balance" size="display" className={className} />
        <MoneyValue cents={-325075} variant="balance" size="display" className={className} />
      </div>
      <div className="flex flex-col items-end gap-1 border-t pt-3">
        {ALIGNMENT.map((cents) => (
          <MoneyValue key={cents} cents={cents} size="display" className={className} />
        ))}
      </div>
      <SpecimenCaption>Alinhados à direita: com algarismos tabulares, 1, 0 e 8 ocupam a mesma largura.</SpecimenCaption>
    </Card>
  );
}

export function StatusSection() {
  return (
    <SpecimenSection
      id="status-e-valores"
      title="Status e valores"
      description="CANCELADA é neutra: cancelamento não é erro. O vermelho de erro fica para falha e saldo negativo."
    >
      <SpecimenBlock title="Status de lançamento">
        <div className={ROW}>
          {STATUS_ORDER.map((status) => (
            <StatusBadge key={status} status={STATUS_LANCAMENTO[status]} />
          ))}
        </div>
      </SpecimenBlock>

      <SpecimenBlock title="Ativo, inativo e saldo negativo">
        <div className={ROW}>
          <StatusBadge status={activeStatusDef(true)} />
          <StatusBadge status={activeStatusDef(false)} />
          <StatusBadge status={SALDO_NEGATIVO} />
        </div>
      </SpecimenBlock>

      <SpecimenBlock title="Badge nos tons">
        <div className={ROW}>
          <Badge variant="neutral">Neutro</Badge>
          <Badge variant="success">Sucesso</Badge>
          <Badge variant="warning">Aviso</Badge>
          <Badge variant="error">Erro</Badge>
          <Badge variant="outline">Contorno</Badge>
        </div>
        <SpecimenCaption>Badge é indicador, não ação: sem hover e sem foco.</SpecimenCaption>
      </SpecimenBlock>

      <SpecimenBlock title="MoneyValue">
        <dl className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-5">
          {[
            { name: "neutral", node: <MoneyValue cents={92050} /> },
            { name: "income", node: <MoneyValue cents={184000} variant="income" /> },
            { name: "expense", node: <MoneyValue cents={92050} variant="expense" /> },
            { name: "balance (positivo)", node: <MoneyValue cents={1847390} variant="balance" /> },
            { name: "balance (negativo)", node: <MoneyValue cents={-325075} variant="balance" /> },
          ].map(({ name, node }) => (
            <div key={name} className="flex flex-col gap-1">
              <dt className="font-mono text-xs text-muted-foreground">{name}</dt>
              <dd className="text-sm">{node}</dd>
            </div>
          ))}
        </dl>
        <SpecimenCaption>O sentido vai em texto: sinal, e ícone com texto oculto no saldo negativo.</SpecimenCaption>
      </SpecimenBlock>

      <SpecimenBlock title="Saldo em destaque: Barlow Condensed e Geist (DD-4)">
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <DisplayComparison family="barlow" />
          <DisplayComparison family="geist" />
        </div>
      </SpecimenBlock>

      <SpecimenBlock title="RoleBadge">
        <div className={ROW}>
          {PAPEIS_EXEMPLO.map((papel) => (
            <RoleBadge key={papel} label={papel} />
          ))}
        </div>
      </SpecimenBlock>
    </SpecimenSection>
  );
}
