"use client";

import { useState } from "react";

import { formatIdentifier } from "@/components/app/format";
import { Button } from "@/components/ui/button";

import { useRegistrarDevolucao } from "./hooks";
import { LancamentoFormDialog } from "./lancamento-form";
import { IdentifierText } from "./identifier";
import { type Lancamento, TIPO_LABEL } from "./labels";

// Devolução de uma receita recebida (FWB-04 AC5): despesa com conta de
// despesa, valores e forma de pagamento, referenciando a receita.

export function DevolucaoAction({ receita }: { receita: Lancamento }) {
  const [open, setOpen] = useState(false);
  const devolucao = useRegistrarDevolucao();
  const identifier = formatIdentifier(TIPO_LABEL[receita.tipo], receita.id);
  return (
    <>
      <Button variant="secondary" onClick={() => setOpen(true)}>
        Registrar devolução
      </Button>
      <LancamentoFormDialog
        open={open}
        onOpenChange={setOpen}
        title="Registrar devolução"
        description={
          <>
            Devolução da <IdentifierText identifier={identifier} />: uma despesa que referencia esta receita
            recebida.
          </>
        }
        defaultValues={{
          tipo: "DESPESA",
          conta_id: "",
          valor_bruto_cents: null,
          taxa_cents: 0,
          forma_pagamento: "",
        }}
        tipoEditable={false}
        submitLabel="Registrar devolução"
        successMessage="Devolução registrada."
        onSubmit={({ conta_id, valor_bruto_cents, taxa_cents, forma_pagamento }) =>
          devolucao.mutateAsync({
            devolucao_de_id: receita.id,
            conta_id,
            valor_bruto_cents,
            taxa_cents,
            forma_pagamento,
          })
        }
      />
    </>
  );
}
