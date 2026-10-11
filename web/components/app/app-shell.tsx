"use client";

import { ChevronDownIcon, MenuIcon } from "lucide-react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { type ReactNode, useState } from "react";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from "@/components/ui/sheet";
import { routes } from "@/lib/session/routes";
import { useSession } from "@/lib/session/session-provider";
import { cn } from "@/lib/utils";

import { isNavSection, type NavEntry, type NavItem, visibleItems } from "./nav";

// Shell das rotas autenticadas (FND-05 AC1, AC2, AC8). A navegação lateral
// aparece a partir do ponto de quebra de tablet (md); abaixo dele vira um
// menu recolhível. O menu da pessoa fica sempre no cabeçalho.

export type AppShellProps = {
  items: readonly NavEntry[];
  children: ReactNode;
};

export function AppShell({ items, children }: AppShellProps) {
  const { context, logout } = useSession();
  const router = useRouter();
  const [menuOpen, setMenuOpen] = useState(false);

  // Com troca de senha obrigatória, nada além de "Sair" (FND-05 AC8).
  const showNav = !!context && !context.must_change_password;
  const entries = showNav ? visibleItems(items, context.permissions) : [];

  return (
    <div className="flex min-h-svh flex-col">
      <header className="sticky top-0 z-40 flex h-14 items-center gap-2 border-b bg-background px-4">
        {showNav ? (
          <Sheet open={menuOpen} onOpenChange={setMenuOpen}>
            <SheetTrigger render={<Button variant="ghost" size="icon" className="md:hidden" />}>
              <MenuIcon aria-hidden />
              <span className="sr-only">Abrir menu</span>
            </SheetTrigger>
            <SheetContent side="left">
              <SheetHeader>
                <SheetTitle>Menu</SheetTitle>
              </SheetHeader>
              <NavList entries={entries} label="Navegação" onNavigate={() => setMenuOpen(false)} />
            </SheetContent>
          </Sheet>
        ) : null}

        {showNav ? (
          <Link href={routes.inicio} className="font-heading font-semibold">
            TJ Platform
          </Link>
        ) : (
          <span className="font-heading font-semibold">TJ Platform</span>
        )}

        <div className="ml-auto">
          <DropdownMenu>
            <DropdownMenuTrigger render={<Button variant="ghost" />}>
              <span className="max-w-48 truncate">{context?.user.name}</span>
              <ChevronDownIcon aria-hidden />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-auto">
              {showNav ? (
                <DropdownMenuItem onClick={() => router.push(routes.contaSenha)}>Trocar senha</DropdownMenuItem>
              ) : null}
              <DropdownMenuItem onClick={() => void logout()}>Sair</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </header>

      <div className="flex flex-1">
        {showNav ? (
          <aside className="hidden w-60 shrink-0 border-r md:block">
            <NavList entries={entries} label="Navegação principal" />
          </aside>
        ) : null}
        <div className="min-w-0 flex-1">{children}</div>
      </div>
    </div>
  );
}

function NavList({
  entries,
  label,
  onNavigate,
}: {
  entries: readonly NavEntry[];
  label: string;
  onNavigate?: () => void;
}) {
  return (
    <nav aria-label={label} className="flex flex-col gap-4 p-3">
      {entries.map((entry) =>
        isNavSection(entry) ? (
          <div key={`secao:${entry.label}`} className="flex flex-col gap-1">
            <p className="px-2 text-xs font-medium text-muted-foreground uppercase">{entry.label}</p>
            <ul className="flex flex-col gap-0.5">
              {entry.items.map((item) => (
                <li key={item.href}>
                  <NavLink item={item} onNavigate={onNavigate} />
                </li>
              ))}
            </ul>
          </div>
        ) : (
          <NavLink key={entry.href} item={entry} onNavigate={onNavigate} />
        ),
      )}
    </nav>
  );
}

function NavLink({ item, onNavigate }: { item: NavItem; onNavigate?: () => void }) {
  const pathname = usePathname();
  const active = pathname === item.href || pathname.startsWith(`${item.href}/`);
  return (
    <Link
      href={item.href}
      aria-current={active ? "page" : undefined}
      onClick={onNavigate}
      className={cn(
        "block rounded-md px-2 py-1.5 text-sm hover:bg-muted",
        active && "bg-muted font-medium",
      )}
    >
      {item.label}
    </Link>
  );
}
