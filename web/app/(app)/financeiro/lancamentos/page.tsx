import type { Metadata } from "next";
import { Suspense } from "react";

import { PageContainer } from "@/components/app/page-container";
import { RequirePermission } from "@/components/app/require-permission";
import { LoadingState } from "@/components/app/states";
import { LancamentosPage } from "@/features/financeiro/lancamentos/lancamentos-page";

export const metadata: Metadata = {
  title: "Lançamentos",
};

// Os filtros vêm da URL (`useSearchParams`), por isso o limite de `Suspense`.
export default function FinanceiroLancamentosPage() {
  return (
    <PageContainer width="wide">
      <RequirePermission permission="financeiro:lancamento:read" mode="page">
        <Suspense fallback={<LoadingState shape="table" />}>
          <LancamentosPage />
        </Suspense>
      </RequirePermission>
    </PageContainer>
  );
}
