"use client";

import { DownloadIcon } from "lucide-react";
import { useState } from "react";

import { type Column, DataList } from "@/components/app/data-list";
import { RowActionButton } from "@/components/app/row-actions";
import { ShowMore } from "@/components/ui/show-more";
import { toast } from "@/components/ui/toast";

import { type Comprovante, COMPROVANTES, LANCAMENTOS } from "./data";
import { LancamentosList } from "./lancamentos-list";
import { SpecimenBlock, SpecimenCaption, SpecimenSection } from "./section";

// Tabela e lista (DSP-05, DSF-06 AC4): a mesma listagem no layout automático
// (tabela a partir de 768 px, sem as colunas de prioridade 3 até 1023 px, e
// lista empilhada abaixo de 768 px) e forçada em `table` e em `list`. A
// tabela pequena de comprovantes continua tabela em qualquer largura.

const PAGE = 8;

const COMPROVANTE_COLUMNS: readonly Column<Comprovante>[] = [
  { id: "arquivo", header: "Arquivo", priority: 1, wrap: "truncate", cell: (row) => row.arquivo },
  { id: "enviado", header: "Enviado em", priority: 1, wrap: "nowrap", cell: (row) => row.enviadoEm },
];

export function ListingSection() {
  const [shown, setShown] = useState(PAGE);
  const first = LANCAMENTOS.slice(0, PAGE);
  return (
    <SpecimenSection
      id="tabela-e-lista"
      title="Tabela e lista"
      description="Redimensione a janela (390, 768, 820, 1024, 1440) para ver o layout automático mudar. Os layouts forçados são a propriedade layout do DataList."
    >
      <SpecimenBlock title='Layout automático (layout="auto") com ShowMore'>
        <LancamentosList label="Lançamentos de exemplo (automático)" rows={LANCAMENTOS.slice(0, shown)} />
        <ShowMore
          shown={shown}
          total={LANCAMENTOS.length}
          noun={["lançamento", "lançamentos"]}
          onShowMore={() => setShown((current) => Math.min(current + PAGE, LANCAMENTOS.length))}
        />
      </SpecimenBlock>

      <SpecimenBlock title='Tabela forçada (layout="table")'>
        {/* Seis colunas forçadas em tabela não cabem em 390 px: só aqui a tabela rola dentro da
            própria região, para a página não rolar na horizontal (o uso real de layout="table" é
            para até quatro colunas curtas). */}
        <div
          role="region"
          tabIndex={0}
          aria-label="Lançamentos de exemplo (tabela), com rolagem horizontal"
          className="relative overflow-x-auto overscroll-x-contain rounded-md outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-ring-offset"
        >
          <LancamentosList label="Lançamentos de exemplo (tabela)" rows={first} layout="table" />
        </div>
        <SpecimenCaption>
          A coluna Conta (prioridade 3) só aparece com tela de 1024 px ou mais e área de conteúdo de pelo menos 896 px. Abaixo de 768 px a moldura rola na horizontal.
        </SpecimenCaption>
      </SpecimenBlock>

      <SpecimenBlock title='Lista empilhada forçada (layout="list")'>
        <LancamentosList label="Lançamentos de exemplo (lista)" rows={first} layout="list" />
      </SpecimenBlock>

      <SpecimenBlock title="Tabela pequena que continua tabela">
        <DataList
          label="Comprovantes de exemplo"
          layout="table"
          columns={COMPROVANTE_COLUMNS}
          rows={COMPROVANTES}
          rowKey={(row) => row.id}
          rowLabel={(row) => `comprovante ${row.arquivo}`}
          primaryAction={(row) => (
            <RowActionButton
              label="Baixar"
              rowLabel={`comprovante ${row.arquivo}`}
              icon={DownloadIcon}
              onClick={() => toast.show("Exemplo: nenhum arquivo é baixado.")}
            />
          )}
        />
        <SpecimenCaption>Poucas colunas curtas: tabela em qualquer largura, também em 390 px (DSP-05 AC8).</SpecimenCaption>
      </SpecimenBlock>
    </SpecimenSection>
  );
}
