import type { Metadata } from "next";

import { PageContainer } from "@/components/app/page-container";
import { RequirePermission } from "@/components/app/require-permission";
import { LancamentoDetail } from "@/features/financeiro/lancamentos/lancamento-detail";

export const metadata: Metadata = {
  title: "Lançamento",
};

export default async function FinanceiroLancamentoPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return (
    <PageContainer>
      <RequirePermission permission="financeiro:lancamento:read" mode="page">
        <LancamentoDetail id={id} />
      </RequirePermission>
    </PageContainer>
  );
}
