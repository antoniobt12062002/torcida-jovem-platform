"use client";

import { KeyRoundIcon, LogOutIcon } from "lucide-react";
import Link from "next/link";
import type { ReactNode } from "react";

import { ShellFrame } from "@/components/app/shell-frame";
import { DropdownMenuItem, DropdownMenuLinkItem, DropdownMenuSeparator } from "@/components/ui/dropdown-menu";
import { toast } from "@/components/ui/toast";

import { NAV_EXEMPLO, PESSOA_EXEMPLO } from "./data";

// O shell real (`ShellFrame`) com navegação, item ativo e menu da pessoa de
// exemplo, sem sessão: nada é lido da API e "Sair" só mostra um aviso.

export const EXAMPLE_PATH = "/design-system/pagina-de-exemplo";

export function ExampleShell({ children }: { children: ReactNode }) {
  return (
    <ShellFrame
      entries={NAV_EXEMPLO}
      currentPath={EXAMPLE_PATH}
      userName={PESSOA_EXEMPLO.nome}
      userEmail={PESSOA_EXEMPLO.email}
      showNav
      userMenu={
        <>
          <DropdownMenuLinkItem render={<Link href="/design-system#shell" />}>
            <KeyRoundIcon aria-hidden="true" />
            Trocar senha
          </DropdownMenuLinkItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={() => toast.show("Exemplo: no espécime não há sessão para encerrar.")}>
            <LogOutIcon aria-hidden="true" />
            Sair
          </DropdownMenuItem>
        </>
      }
    >
      {children}
    </ShellFrame>
  );
}
