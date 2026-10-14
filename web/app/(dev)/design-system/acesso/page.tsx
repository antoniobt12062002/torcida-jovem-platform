import type { Metadata } from "next";

import { AccessLayout } from "@/components/app/access-layout";

import { notFoundInProduction } from "../dev-only";
import { ExampleLoginForm } from "./example-login-form";

// Tela de acesso de exemplo do espécime (DSS-03): o AccessLayout real com um
// formulário de login sem envio. Em produção responde como não encontrada.

export function generateMetadata(): Metadata {
  notFoundInProduction();
  return { title: "Espécime: acesso", robots: { index: false, follow: false } };
}

export default function DesignSystemAccessPage() {
  notFoundInProduction();
  return (
    <AccessLayout>
      <ExampleLoginForm />
    </AccessLayout>
  );
}
