"use client";

import { KeyRoundIcon, LogOutIcon } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";

import { DropdownMenuItem, DropdownMenuLinkItem, DropdownMenuSeparator } from "@/components/ui/dropdown-menu";
import { routes } from "@/lib/session/routes";
import { useSession } from "@/lib/session/session-provider";

import { type NavEntry, visibleItems } from "./nav";
import { ShellFrame } from "./shell-frame";

// Shell das rotas autenticadas (FND-05 AC1, AC2, AC8; DSS-01, DSS-02). Lê a
// sessão e as permissões e entrega tudo pronto ao `ShellFrame`, que é só
// apresentação. O menu da pessoa fica sempre no cabeçalho.

export type AppShellProps = {
  items: readonly NavEntry[];
  children: ReactNode;
};

export function AppShell({ items, children }: AppShellProps) {
  const { context, logout } = useSession();
  const pathname = usePathname();

  // Com troca de senha obrigatória, nada além de "Sair" (FND-05 AC8).
  const showNav = !!context && !context.must_change_password;
  const entries = showNav ? visibleItems(items, context.permissions) : [];

  return (
    <ShellFrame
      entries={entries}
      currentPath={pathname}
      userName={context?.user.name ?? ""}
      userEmail={context?.user.email}
      showNav={showNav}
      userMenu={
        <>
          {showNav ? (
            <>
              {/* Navegação é link, não botão com `onClick` (DSS-01 AC7). */}
              <DropdownMenuLinkItem render={<Link href={routes.contaSenha} />}>
                <KeyRoundIcon aria-hidden />
                Trocar senha
              </DropdownMenuLinkItem>
              <DropdownMenuSeparator />
            </>
          ) : null}
          <DropdownMenuItem nativeButton render={<button type="button" />} onClick={() => void logout()}>
            <LogOutIcon aria-hidden />
            Sair
          </DropdownMenuItem>
        </>
      }
    >
      {children}
    </ShellFrame>
  );
}
