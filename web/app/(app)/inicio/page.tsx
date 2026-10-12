import type { Metadata } from "next";

import { PageContainer } from "@/components/app/page-container";

import { Atalhos } from "./atalhos";

export const metadata: Metadata = {
  title: "Início",
};

// Página inicial autenticada (INT-01 AC4, DSS-06): atalhos para as áreas que
// a pessoa pode ver, gerados da mesma lista da navegação. Sem indicadores,
// gráficos nem dados de exemplo (WDS-D-010).
export default function InicioPage() {
  return (
    <PageContainer>
      <Atalhos />
    </PageContainer>
  );
}
