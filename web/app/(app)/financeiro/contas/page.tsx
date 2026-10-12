import type { Metadata } from "next";

import { PageContainer } from "@/components/app/page-container";
import { RequirePermission } from "@/components/app/require-permission";
import { ContasPage } from "@/features/financeiro/contas/contas-page";

export const metadata: Metadata = {
  title: "Plano de contas",
};

export default function FinanceiroContasPage() {
  return (
    <PageContainer width="default">
      <RequirePermission permission="financeiro:conta:read" mode="page">
        <ContasPage />
      </RequirePermission>
    </PageContainer>
  );
}
