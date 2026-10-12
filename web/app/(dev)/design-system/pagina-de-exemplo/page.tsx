import type { Metadata } from "next";

import { notFoundInProduction } from "../dev-only";
import { ExamplePage } from "./example-page";

// Página de exemplo do espécime: shell, trilha, cabeçalho de página, barra de
// filtros e listagem juntos, como numa tela real. Em produção responde como
// não encontrada.

export function generateMetadata(): Metadata {
  notFoundInProduction();
  return { title: "Espécime: página de exemplo", robots: { index: false, follow: false } };
}

export default function DesignSystemExamplePage() {
  notFoundInProduction();
  return <ExamplePage />;
}
