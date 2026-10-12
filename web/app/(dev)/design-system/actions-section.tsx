"use client";

import { DownloadIcon, EllipsisVerticalIcon, PlusIcon, SearchIcon } from "lucide-react";
import Link from "next/link";

import { Button, buttonVariants } from "@/components/ui/button";
import { IconButton } from "@/components/ui/icon-button";

import { SpecimenBlock, SpecimenCaption, SpecimenSection } from "./section";

// Ações (DSC-02): Button em todas as variantes e tamanhos, em repouso,
// desabilitado e em andamento; IconButton; link com aparência de botão. O foco
// visível aparece ao navegar com Tab.

const VARIANTS = [
  { variant: "primary", label: "Salvar lançamento" },
  { variant: "secondary", label: "Exportar" },
  { variant: "ghost", label: "Limpar filtros" },
  { variant: "destructive", label: "Excluir rascunho" },
  { variant: "link", label: "Ver histórico" },
] as const;

const SIZES = ["sm", "md", "lg"] as const;

const ROW = "flex flex-wrap items-center gap-3";

export function ActionsSection() {
  return (
    <SpecimenSection
      id="acoes"
      title="Ações"
      description="Uma ação primária por área. Em andamento: desabilitado, aria-busy e rótulo terminado em reticências. Use Tab para ver o foco."
    >
      {VARIANTS.map(({ variant, label }) => (
        <SpecimenBlock key={variant} title={`Button variant="${variant}"`}>
          <div className="flex flex-col gap-3">
            {SIZES.map((size) => (
              <div key={size} className={ROW}>
                <span className="w-16 shrink-0 font-mono text-xs text-muted-foreground">{size}</span>
                <Button variant={variant} size={size}>
                  {label}
                </Button>
                <Button variant={variant} size={size} disabled>
                  {label}
                </Button>
                <Button variant={variant} size={size} loading loadingLabel="Salvando…">
                  {label}
                </Button>
              </div>
            ))}
          </div>
          <SpecimenCaption>Por linha: repouso, desabilitado e em andamento.</SpecimenCaption>
        </SpecimenBlock>
      ))}

      <SpecimenBlock title="Button com ícone">
        <div className={ROW}>
          <Button>
            <PlusIcon aria-hidden="true" />
            Novo lançamento
          </Button>
          <Button variant="secondary">
            <DownloadIcon aria-hidden="true" />
            Baixar comprovante
          </Button>
        </div>
      </SpecimenBlock>

      <SpecimenBlock title="IconButton">
        <div className="flex flex-col gap-3">
          {(["ghost", "secondary"] as const).map((variant) => (
            <div key={variant} className={ROW}>
              <span className="w-16 shrink-0 font-mono text-xs text-muted-foreground">{variant}</span>
              {SIZES.map((size) => (
                <IconButton key={size} label={`Mais ações (${size})`} variant={variant} size={size}>
                  <EllipsisVerticalIcon aria-hidden="true" />
                </IconButton>
              ))}
              <IconButton label="Buscar (desabilitado)" variant={variant} disabled>
                <SearchIcon aria-hidden="true" />
              </IconButton>
            </div>
          ))}
        </div>
        <SpecimenCaption>sm, md e lg, depois desabilitado. O nome acessível é obrigatório.</SpecimenCaption>
      </SpecimenBlock>

      <SpecimenBlock title="Link com aparência de botão">
        <div className={ROW}>
          <Link href="#acoes" className={buttonVariants({ variant: "secondary" })}>
            Voltar para a lista
          </Link>
          <Link href="#acoes" className={buttonVariants({ variant: "primary" })}>
            Ir para o saldo
          </Link>
        </div>
        <SpecimenCaption>Navegação é link (Link com buttonVariants), não botão com onClick.</SpecimenCaption>
      </SpecimenBlock>
    </SpecimenSection>
  );
}
