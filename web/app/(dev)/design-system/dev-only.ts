import { notFound } from "next/navigation";

// O espécime é uma referência interna (DSF-06 AC5, WDS-D-017): em produção
// toda página da rota responde como não encontrada, antes de renderizar
// qualquer conteúdo. A revisão do mantenedor é feita com `pnpm dev`.

export function notFoundInProduction() {
  if (process.env.NODE_ENV === "production") notFound();
}
