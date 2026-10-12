import type { Metadata } from "next";
import { Suspense } from "react";

import { ChangePasswordForm } from "@/features/identity/acesso/change-password-form";

export const metadata: Metadata = {
  title: "Trocar senha",
};

export default function ContaSenhaPage() {
  // O formulário lê `next` da URL (useSearchParams): fica num limite de
  // Suspense. Ele escolhe o próprio contêiner: layout de acesso na troca
  // obrigatória, página estreita do shell na voluntária.
  return (
    <Suspense>
      <ChangePasswordForm />
    </Suspense>
  );
}
