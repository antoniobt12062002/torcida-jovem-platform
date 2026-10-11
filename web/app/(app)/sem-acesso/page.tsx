import type { Metadata } from "next";
import Link from "next/link";

import { routes } from "@/lib/session/routes";

export const metadata: Metadata = {
  title: "Sem acesso",
};

// FND-04 AC4: área sem a permissão de leitura.
export default function SemAcessoPage() {
  return (
    <main className="mx-auto flex w-full max-w-md flex-1 flex-col justify-center gap-3 px-4 py-12">
      <h1 className="text-xl font-semibold">Sem acesso</h1>
      <p className="text-sm text-muted-foreground">
        Você não tem permissão para acessar esta área. Se precisar dela, fale com a administração.
      </p>
      <Link
        href={routes.inicio}
        className="text-sm underline underline-offset-4 hover:text-foreground"
      >
        Voltar ao início
      </Link>
    </main>
  );
}
