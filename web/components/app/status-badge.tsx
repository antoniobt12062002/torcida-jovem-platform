import { CircleCheckIcon, CircleSlashIcon, type LucideIcon } from "lucide-react";

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

export type StatusBadgeProps = {
  /** Definição da feature dona. */
  status: StatusDef;
  className?: string;
};

export function StatusBadge({ status, className }: StatusBadgeProps) {
  const { label, tone, icon: Icon } = status;
  return (
    <Badge variant={tone} data-tone={tone} className={className}>
      <Icon aria-hidden="true" />
      {label}
    </Badge>
  );
}
