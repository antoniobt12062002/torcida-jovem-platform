"use client";

import { ChevronDownIcon, MenuIcon, XIcon } from "lucide-react";
import Link from "next/link";
import { type CSSProperties, type MouseEvent, type ReactNode, useId, useRef, useState } from "react";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { IconButton } from "@/components/ui/icon-button";
import { Sheet, SheetClose, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from "@/components/ui/sheet";
import { DisplayTitle, Eyebrow } from "@/components/ui/typography";
import { BREAKPOINT_LG, useMediaQuery } from "@/lib/responsive";
import { routes } from "@/lib/session/routes";
import { cn } from "@/lib/utils";

import { BRAND_NAME, BrandMark } from "./brand";
import { isNavSection, type NavEntry, type NavItem } from "./nav";

// Moldura das rotas autenticadas, só apresentação (DSS-01, DSS-02): não lê
// sessão nem API, para o espécime mostrar o shell real com dados de exemplo
// (WDS-D-017). Quem conhece a sessão é o `AppShell`.
//
// Cabeçalho e navegação usam os tokens do shell escuro; o conteúdo, os do
// conteúdo claro. A navegação fica fixa a partir de 1024 px e numa gaveta
// abaixo disso (WDS-D-019), trocadas só por CSS.

export type ShellFrameProps = {
  /** Itens da navegação, já filtrados por permissão. */
  entries: readonly NavEntry[];
  /** Caminho atual, para marcar o item ativo. */
  currentPath: string;
  userName: string;
  userEmail?: string;
  /** Itens do menu da pessoa: `DropdownMenuLinkItem` e `DropdownMenuItem`. */
  userMenu: ReactNode;
  /** Falso durante a troca obrigatória de senha: sem navegação nem links. */
  showNav: boolean;
  children: ReactNode;
};

/** Alvo do link de pular: o `main` do `PageContainer`. */
const SKIP_TARGET_ID = "conteudo";

// Botões do kit sobre o fundo escuro: mesmas medidas, cores do shell.
const ON_SHELL =
  "text-sidebar-foreground hover:bg-sidebar-accent active:bg-sidebar-accent aria-expanded:bg-sidebar-accent focus-visible:ring-sidebar-ring focus-visible:ring-offset-sidebar";

// Áreas seguras do aparelho (DSS-02 AC6). Valores de `env()` não têm token.
const HEADER_SAFE_AREA: CSSProperties = {
  paddingTop: "env(safe-area-inset-top)",
  paddingLeft: "env(safe-area-inset-left)",
  paddingRight: "env(safe-area-inset-right)",
};
const DRAWER_SAFE_AREA: CSSProperties = {
  paddingTop: "env(safe-area-inset-top)",
  paddingBottom: "env(safe-area-inset-bottom)",
  paddingLeft: "env(safe-area-inset-left)",
};
// A navegação fixa acompanha a rolagem da página e rola por conta própria.
const RAIL_VIEWPORT: CSSProperties = {
  top: "calc(var(--shell-header-height) + env(safe-area-inset-top))",
  maxHeight: "calc(100svh - var(--shell-header-height) - env(safe-area-inset-top))",
  paddingLeft: "env(safe-area-inset-left)",
};

export function ShellFrame({ entries, currentPath, userName, userEmail, userMenu, showNav, children }: ShellFrameProps) {
  const contentRef = useRef<HTMLDivElement>(null);
  const [menuOpen, setMenuOpen] = useState(false);

  // Ao chegar a 1024 px a navegação fixa aparece: a gaveta aberta fecha.
  const wide = useMediaQuery(`(min-width: ${BREAKPOINT_LG}px)`);
  const [wasWide, setWasWide] = useState(wide);
  if (wide !== wasWide) {
    setWasWide(wide);
    if (wide) setMenuOpen(false);
  }

  // As páginas migradas trazem `main#conteudo` (`PageContainer`). Até todas
  // migrarem, o alvo é resolvido na hora: o `main` da página, ou a própria
  // área de conteúdo. O shell nunca cria outro `main` nem repete o `id`.
  function skipToContent(event: MouseEvent<HTMLAnchorElement>) {
    const content = contentRef.current;
    if (!content) return;
    const target =
      content.querySelector<HTMLElement>(`#${SKIP_TARGET_ID}`) ?? content.querySelector<HTMLElement>("main") ?? content;
    event.preventDefault();
    if (!target.hasAttribute("tabindex")) target.setAttribute("tabindex", "-1");
    target.focus();
  }

  const brand = (
    <>
      <BrandMark size={32} priority />
      <DisplayTitle aria-hidden className="hidden text-xl leading-6 min-[400px]:block">
        {BRAND_NAME}
      </DisplayTitle>
    </>
  );

  return (
    <div data-slot="shell" className="flex min-h-svh flex-col">
      {showNav ? (
        <a
          href={`#${SKIP_TARGET_ID}`}
          onClick={skipToContent}
          className="fixed top-0 left-0 z-50 flex h-control-lg -translate-y-full items-center rounded-br-md bg-background px-4 text-sm font-medium text-foreground underline underline-offset-4 outline-none focus:translate-y-0 focus-visible:ring-2 focus-visible:ring-sidebar-ring"
        >
          Pular para o conteúdo
        </a>
      ) : null}

      <header
        className="sticky top-0 z-40 border-b border-sidebar-border bg-sidebar text-sidebar-foreground"
        style={HEADER_SAFE_AREA}
      >
        <div className="flex h-shell-header items-center gap-2 px-4 lg:px-6">
          {showNav ? (
            <Sheet open={menuOpen} onOpenChange={setMenuOpen}>
              <SheetTrigger
                render={
                  <IconButton label="Abrir menu" size="lg" className={cn(ON_SHELL, "lg:hidden")}>
                    <MenuIcon aria-hidden />
                  </IconButton>
                }
              />
              <SheetContent
                side="left"
                showCloseButton={false}
                className="w-70 max-w-full gap-0 border-sidebar-border bg-sidebar text-sidebar-foreground"
                style={DRAWER_SAFE_AREA}
              >
                <SheetHeader className="flex-row items-center justify-between py-2 pr-2">
                  <SheetTitle className="text-sidebar-foreground">Menu</SheetTitle>
                  <SheetClose
                    render={
                      <IconButton label="Fechar" size="lg" className={ON_SHELL}>
                        <XIcon aria-hidden />
                      </IconButton>
                    }
                  />
                </SheetHeader>
                <NavList entries={entries} label="Menu" currentPath={currentPath} onNavigate={() => setMenuOpen(false)} />
              </SheetContent>
            </Sheet>
          ) : null}

          {showNav ? (
            <Link
              href={routes.inicio}
              className="flex min-h-touch min-w-touch touch-manipulation items-center gap-1 rounded-md outline-none transition-colors duration-fast hover:bg-sidebar-accent focus-visible:ring-2 focus-visible:ring-sidebar-ring"
            >
              {brand}
            </Link>
          ) : (
            <div className="flex items-center gap-1">{brand}</div>
          )}

          <div className="ml-auto min-w-0">
            <DropdownMenu>
              <DropdownMenuTrigger render={<Button variant="ghost" className={cn(ON_SHELL, "max-w-full px-3")} />}>
                <span className="max-w-48 truncate">{userName}</span>
                <ChevronDownIcon aria-hidden />
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-auto min-w-56">
                {/* Nome e e-mail rotulam o grupo dos itens: estrutura válida dentro de `role="menu"`. */}
                <DropdownMenuGroup>
                  <DropdownMenuLabel className="flex flex-col px-2 py-1.5 font-normal text-popover-foreground">
                    <span className="text-sm font-semibold wrap-break-word">{userName}</span>
                    {userEmail ? (
                      <span className="text-xs wrap-break-word text-muted-foreground">{userEmail}</span>
                    ) : null}
                  </DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  {userMenu}
                </DropdownMenuGroup>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </header>

      <div className="flex flex-1">
        {showNav ? (
          <div
            data-slot="shell-nav"
            className="hidden w-shell-nav shrink-0 border-r border-sidebar-border bg-sidebar text-sidebar-foreground lg:block"
          >
            <div className="sticky overflow-y-auto overscroll-contain" style={RAIL_VIEWPORT}>
              <NavList entries={entries} label="Navegação principal" currentPath={currentPath} />
            </div>
          </div>
        ) : null}
        <div ref={contentRef} data-slot="shell-content" className="flex min-w-0 flex-1 flex-col bg-background outline-none">
          {children}
        </div>
      </div>
    </div>
  );
}

function NavList({
  entries,
  label,
  currentPath,
  onNavigate,
}: {
  entries: readonly NavEntry[];
  label: string;
  currentPath: string;
  onNavigate?: () => void;
}) {
  // A gaveta e a navegação fixa existem juntas: cada lista tem o próprio prefixo.
  const idPrefix = useId();
  return (
    <nav aria-label={label} className="flex flex-col gap-6 p-3">
      {entries.map((entry, index) =>
        isNavSection(entry) ? (
          <div key={`secao:${entry.label}`} className="flex flex-col gap-1">
            <Eyebrow id={`${idPrefix}-grupo-${index}`} className="px-3 py-1 text-sidebar-muted-foreground">
              {entry.label}
            </Eyebrow>
            <ul aria-labelledby={`${idPrefix}-grupo-${index}`} className="flex flex-col gap-0.5">
              {entry.items.map((item) => (
                <li key={item.href}>
                  <NavLink item={item} currentPath={currentPath} onNavigate={onNavigate} />
                </li>
              ))}
            </ul>
          </div>
        ) : (
          <NavLink key={entry.href} item={entry} currentPath={currentPath} onNavigate={onNavigate} />
        ),
      )}
    </nav>
  );
}

function NavLink({ item, currentPath, onNavigate }: { item: NavItem; currentPath: string; onNavigate?: () => void }) {
  const active = currentPath === item.href || currentPath.startsWith(`${item.href}/`);
  const Icon = item.icon;
  return (
    <Link
      href={item.href}
      aria-current={active ? "page" : undefined}
      onClick={onNavigate}
      className={cn(
        "relative flex min-h-control-md touch-manipulation items-center gap-3 rounded-md px-3 text-sm text-sidebar-foreground outline-none transition-colors duration-fast focus-visible:ring-2 focus-visible:ring-sidebar-ring",
        // Ativo: peso e barra, sempre sobre `sidebar` (não troca de fundo no hover).
        active
          ? "font-semibold before:absolute before:inset-y-1.5 before:left-0 before:w-0.75 before:bg-sidebar-primary"
          : "font-normal hover:bg-sidebar-accent",
      )}
    >
      {/* Sem ícone, o espaço dele é mantido para os rótulos ficarem alinhados. */}
      {Icon ? <Icon aria-hidden className="size-5 shrink-0" /> : <span aria-hidden className="size-5 shrink-0" />}
      {item.label}
    </Link>
  );
}
