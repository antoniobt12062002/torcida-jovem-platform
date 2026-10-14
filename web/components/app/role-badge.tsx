import { Badge } from "@/components/ui/badge";

// Selo de papel (DSC-06 AC5): o nome do papel em português, em tom neutro. O
// mapa de rótulos é da feature de usuários; o componente recebe o rótulo pronto.

export type RoleBadgeProps = { label: string; className?: string };

export function RoleBadge({ label, className }: RoleBadgeProps) {
  return (
    <Badge variant="outline" className={className}>
      {label}
    </Badge>
  );
}
