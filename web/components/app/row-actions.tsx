"use client";

import { EllipsisVerticalIcon, type LucideIcon } from "lucide-react";
import Link from "next/link";
import type { ReactNode } from "react";

import { Button, buttonVariants } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLinkItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { IconButton } from "@/components/ui/icon-button";

// Ações de linha (DSP-05 AC10 e AC11): uma ação visível e as demais num menu.
// Todo nome acessível começa pelo texto visível e termina identificando a
// linha (WCAG 2.5.3): "Devolver" vira "Devolver {rowLabel}", e o menu é
// "Ações de {rowLabel}".

export type RowAction = {
  label: string;
  icon?: LucideIcon;
  onSelect?: () => void;
  /** Com destino, o item é um link. */
  href?: string;
  /** Vai por último no menu, depois de um separador. */
  destructive?: boolean;
  disabled?: boolean;
};

/** Nome acessível de uma ação de linha: o texto visível seguido da linha. */
export function rowActionName(label: string, rowLabel: string) {
  return `${label} ${rowLabel}`;
}

export type RowActionButtonProps = {
  /** Texto visível, igual em todas as linhas. */
  label: string;
  /** Identifica a linha; entra só no nome acessível. */
  rowLabel: string;
  icon?: LucideIcon;
  onClick?: () => void;
  /** Com destino, a ação é um link. */
  href?: string;
  disabled?: boolean;
};

/** A ação visível da linha, com o nome acessível que identifica a linha. */
export function RowActionButton({ label, rowLabel, icon: Icon, onClick, href, disabled }: RowActionButtonProps) {
  const content = (
    <>
      {Icon ? <Icon aria-hidden /> : null}
      {label}
    </>
  );
  const name = rowActionName(label, rowLabel);
  if (href !== undefined && !disabled) {
    return (
      <Link href={href} aria-label={name} className={buttonVariants({ variant: "secondary", size: "sm" })}>
        {content}
      </Link>
    );
  }
  return (
    <Button variant="secondary" size="sm" aria-label={name} onClick={onClick} disabled={disabled}>
      {content}
    </Button>
  );
}

export type RowActionsProps = {
  /** Identifica a linha, por exemplo "lançamento Receita fb6e3d8f". */
  rowLabel: string;
  /** A ação visível; use `RowActionButton` para o nome identificar a linha. */
  primary?: ReactNode;
  /** As demais ações, reunidas no menu. */
  actions?: readonly RowAction[];
};

export function RowActions({ rowLabel, primary, actions = [] }: RowActionsProps) {
  if (!primary && actions.length === 0) return null;
  const regular = actions.filter((action) => !action.destructive);
  const destructive = actions.filter((action) => action.destructive);
  return (
    <div data-slot="row-actions" className="flex items-center justify-end gap-1">
      {primary}
      {actions.length > 0 ? (
        <DropdownMenu>
          <DropdownMenuTrigger
            render={
              <IconButton label={`Ações de ${rowLabel}`}>
                <EllipsisVerticalIcon aria-hidden />
              </IconButton>
            }
          />
          <DropdownMenuContent align="end" className="w-auto">
            {regular.map((action) => (
              <MenuAction key={action.label} action={action} rowLabel={rowLabel} />
            ))}
            {regular.length > 0 && destructive.length > 0 ? <DropdownMenuSeparator /> : null}
            {destructive.map((action) => (
              <MenuAction key={action.label} action={action} rowLabel={rowLabel} />
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      ) : null}
    </div>
  );
}

function MenuAction({ action, rowLabel }: { action: RowAction; rowLabel: string }) {
  const { label, icon: Icon, href, onSelect, destructive, disabled } = action;
  const name = rowActionName(label, rowLabel);
  const content = (
    <>
      {Icon ? <Icon aria-hidden /> : null}
      {label}
    </>
  );
  if (href !== undefined && !disabled) {
    return (
      <DropdownMenuLinkItem aria-label={name} render={<Link href={href} />}>
        {content}
      </DropdownMenuLinkItem>
    );
  }
  return (
    <DropdownMenuItem
      aria-label={name}
      variant={destructive ? "destructive" : "default"}
      disabled={disabled}
      onClick={onSelect}
    >
      {content}
    </DropdownMenuItem>
  );
}
