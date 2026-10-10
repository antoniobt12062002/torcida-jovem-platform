import type { Metadata } from "next";

import { ResetPasswordForm } from "@/features/identity/acesso/reset-password-form";

export const metadata: Metadata = {
  title: "Redefinir senha",
};

export default function RedefinirSenhaPage() {
  return <ResetPasswordForm />;
}
