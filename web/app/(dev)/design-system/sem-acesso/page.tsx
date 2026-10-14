import type { Metadata } from "next";

import { notFoundInProduction } from "../dev-only";
import { NoAccessExample } from "./no-access-example";

// Estado "Sem acesso" do espécime, dentro do shell: traz o único `h1` da
// página, por isso não fica na página principal. Em produção responde como
// não encontrada.

export function generateMetadata(): Metadata {
  notFoundInProduction();
  return { title: "Espécime: sem acesso", robots: { index: false, follow: false } };
}

export default function DesignSystemNoAccessPage() {
  notFoundInProduction();
  return <NoAccessExample />;
}
