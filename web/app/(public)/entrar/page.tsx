import type { Metadata } from "next";
import { Suspense } from "react";

import { LoginForm } from "./login-form";

export const metadata: Metadata = {
  title: "Entrar",
};

export default function EntrarPage() {
  // O formulário lê `next` da URL (useSearchParams): fica num limite de
  // Suspense para o resto da página continuar pré-renderizado.
  return (
    <Suspense>
      <LoginForm />
    </Suspense>
  );
}
