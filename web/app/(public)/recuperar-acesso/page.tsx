import type { Metadata } from "next";

import { RequestResetForm } from "@/features/identity/acesso/request-reset-form";

export const metadata: Metadata = {
  title: "Recuperar acesso",
};

export default function RecuperarAcessoPage() {
  return <RequestResetForm />;
}
