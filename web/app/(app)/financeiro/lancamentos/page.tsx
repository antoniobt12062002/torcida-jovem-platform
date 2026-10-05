import type { Metadata } from "next";

import { RequirePermission } from "@/components/app/require-permission";
import { LancamentosPage } from "@/features/financeiro/lancamentos/lancamentos-page";

export const metadata: Metadata = {
  title: "Lançamentos",
};

export default function FinanceiroLancamentosPage() {
  return (
    <main className="mx-auto flex w-full max-w-6xl flex-1 flex-col px-4 py-8">
      <RequirePermission permission="financeiro:lancamento:read" mode="page">
        <LancamentosPage />
      </RequirePermission>
    </main>
  );
}
