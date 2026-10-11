import type { Metadata } from "next";
import { Suspense } from "react";

import { ChangePasswordForm } from "@/features/identity/acesso/change-password-form";

export const metadata: Metadata = {
  title: "Trocar senha",
};

export default function ContaSenhaPage() {
  // O formulário lê `next` da URL (useSearchParams): fica num limite de Suspense.
  return (
    <main className="mx-auto flex w-full max-w-sm flex-1 flex-col justify-center px-4 py-12">
      <Suspense>
        <ChangePasswordForm />
      </Suspense>
    </main>
  );
}
