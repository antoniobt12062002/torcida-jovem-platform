import type { ReactNode } from "react";

import { BrandMark } from "./brand";

// Layout das telas de acesso (DSS-03): login, recuperação, redefinição e troca
// obrigatória de senha. A partir de 1024 px, painel de marca à esquerda e
// formulário à direita; abaixo disso, faixa de marca no topo e formulário em
// coluna única. A troca é só CSS. Sem ilustração nem fotografia, e sem repetir
// "Desde 2013", que fica só dentro do escudo (WDS-D-027).

export type AccessLayoutProps = {
  /** O título (`h1`) e o formulário da tela. */
  children: ReactNode;
  /** Ação no canto, fora do formulário: "Sair" na troca obrigatória de senha. */
  action?: ReactNode;
};

export function AccessLayout({ children, action }: AccessLayoutProps) {
  return (
    <div data-slot="access-layout" className="flex min-h-svh flex-1 flex-col lg:flex-row">
      <div
        data-slot="access-brand"
        className="flex shrink-0 items-center justify-center bg-sidebar px-4 py-6 text-sidebar-foreground lg:w-2/5 lg:py-12"
      >
        <BrandMark size={64} withName direction="column" className="lg:hidden" />
        <BrandMark size={96} withName direction="column" className="hidden lg:inline-flex" />
      </div>
      <main
        id="conteudo"
        tabIndex={-1}
        className="relative flex flex-1 flex-col items-center gap-6 bg-background px-4 py-8 outline-none lg:justify-center lg:py-12"
      >
        {action ? <div className="flex w-full justify-end lg:absolute lg:top-4 lg:right-4 lg:w-auto">{action}</div> : null}
        <div className="w-full max-w-100">{children}</div>
      </main>
    </div>
  );
}
