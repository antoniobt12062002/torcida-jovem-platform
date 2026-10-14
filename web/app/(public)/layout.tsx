import type { ReactNode } from "react";

import { AccessLayout } from "@/components/app/access-layout";

// Rotas sem sessão (login e recuperação de acesso), no layout de acesso
// (DSS-03): escudo e "Torcida Jovem"; o nome do sistema fica só no título da aba
// (WDS-D-020).
export default function PublicLayout({ children }: { children: ReactNode }) {
  return <AccessLayout>{children}</AccessLayout>;
}
