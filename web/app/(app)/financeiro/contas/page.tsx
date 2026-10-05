import type { Metadata } from "next";

import { RequirePermission } from "@/components/app/require-permission";
import { ContasPage, PERMISSION } from "@/features/financeiro/contas/contas-page";

export const metadata: Metadata = {
  title: "Plano de contas",
};

export default function FinanceiroContasPage() {
  return (
    <main className="mx-auto flex w-full max-w-5xl flex-1 flex-col px-4 py-8">
      <RequirePermission permission={PERMISSION.read} mode="page">
        <ContasPage />
      </RequirePermission>
    </main>
  );
}
