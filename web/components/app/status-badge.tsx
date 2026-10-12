import {
  ArrowDownToLineIcon,
  ArrowUpFromLineIcon,
  BanIcon,
  CircleCheckIcon,
  CircleDashedIcon,
  CircleSlashIcon,
  type LucideIcon,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";

// Indicador de status (DSC-05 AC2 a AC4): tom, ícone e rótulo, nunca só cor.
// O componente não conhece os status de nenhum módulo: cada feature exporta o
// próprio mapa de `StatusDef`. "Ativo e inativo" é comum aos módulos e fica aqui.

export type Tone = "neutral" | "success" | "warning" | "error";

export type StatusDef = { label: string; tone: Tone; icon: LucideIcon };

const ACTIVE: StatusDef = { label: "Ativo", tone: "success", icon: CircleCheckIcon };
const INACTIVE: StatusDef = { label: "Inativo", tone: "neutral", icon: CircleSlashIcon };

/** Definição pronta para o booleano `ativo`/`active` dos recursos. */
export function activeStatusDef(active: boolean): StatusDef {
  return active ? ACTIVE : INACTIVE;
}

// Alias de compatibilidade: removido em DS-INT. Status como texto, com o mapa
// fixo abaixo; as telas migradas passam um `StatusDef` da própria feature.
export type Status = "CRIADA" | "RECEBIDA" | "PAGA" | "CANCELADA" | "ativo" | "inativo";

// Alias de compatibilidade: removido em DS-INT. Mapa de DD-3 (WDS-D-026):
// `CANCELADA` é neutra, porque cancelamento não é erro.
const LEGACY: Record<Status, StatusDef> = {
  CRIADA: { label: "Criada", tone: "neutral", icon: CircleDashedIcon },
  RECEBIDA: { label: "Recebida", tone: "success", icon: ArrowDownToLineIcon },
  PAGA: { label: "Paga", tone: "success", icon: ArrowUpFromLineIcon },
  CANCELADA: { label: "Cancelada", tone: "neutral", icon: BanIcon },
  ativo: ACTIVE,
  inativo: INACTIVE,
};

// Alias de compatibilidade: removido em DS-INT (use `activeStatusDef`).
/** Converte o booleano `ativo`/`active` dos recursos em status. */
export function activeStatus(active: boolean): Status {
  return active ? "ativo" : "inativo";
}

export type StatusBadgeProps = {
  /** Definição da feature dona. O texto é o alias de compatibilidade, removido em DS-INT. */
  status: StatusDef | Status;
  className?: string;
};

export function StatusBadge({ status, className }: StatusBadgeProps) {
  const { label, tone, icon: Icon } = typeof status === "string" ? LEGACY[status] : status;
  return (
    <Badge variant={tone} data-tone={tone} className={className}>
      <Icon aria-hidden="true" />
      {label}
    </Badge>
  );
}
