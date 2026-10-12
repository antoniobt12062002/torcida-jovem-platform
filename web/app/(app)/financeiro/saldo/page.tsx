import type { Metadata } from "next";

import { PageContainer } from "@/components/app/page-container";
import { RequirePermission } from "@/components/app/require-permission";
import { SaldoPage } from "@/features/financeiro/saldo/saldo-page";

export const metadata: Metadata = {
  title: "Saldo",
};

export default function FinanceiroSaldoPage() {
  return (
    <PageContainer width="narrow">
      <RequirePermission permission="financeiro:saldo:read" mode="page">
        <SaldoPage />
      </RequirePermission>
    </PageContainer>
  );
}
