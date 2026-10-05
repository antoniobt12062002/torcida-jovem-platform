import { Badge } from "@/components/ui/badge";
import type { components } from "@/lib/api/financeiro";

// Badge de status com rótulo em português.

export type Status = components["schemas"]["StatusLancamento"] | "ativo" | "inativo";

type BadgeVariant = "default" | "secondary" | "destructive" | "outline";

const STATUS: Record<Status, { label: string; variant: BadgeVariant }> = {
  CRIADA: { label: "Criada", variant: "outline" },
  RECEBIDA: { label: "Recebida", variant: "default" },
  PAGA: { label: "Paga", variant: "default" },
  CANCELADA: { label: "Cancelada", variant: "destructive" },
  ativo: { label: "Ativo", variant: "secondary" },
  inativo: { label: "Inativo", variant: "outline" },
};

/** Converte o booleano `ativo`/`active` dos recursos em status. */
export function activeStatus(active: boolean): Status {
  return active ? "ativo" : "inativo";
}

export function StatusBadge({ status, className }: { status: Status; className?: string }) {
  const { label, variant } = STATUS[status];
  return (
    <Badge variant={variant} className={className}>
      {label}
    </Badge>
  );
}
