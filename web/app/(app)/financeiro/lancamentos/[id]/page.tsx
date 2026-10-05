import type { Metadata } from "next";

import { RequirePermission } from "@/components/app/require-permission";
import { LancamentoDetail } from "@/features/financeiro/lancamentos/lancamento-detail";

export const metadata: Metadata = {
  title: "Lançamento",
};

export default async function FinanceiroLancamentoPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return (
    <main className="mx-auto flex w-full max-w-4xl flex-1 flex-col px-4 py-8">
      <RequirePermission permission="financeiro:lancamento:read" mode="page">
        <LancamentoDetail id={id} />
      </RequirePermission>
    </main>
  );
}
