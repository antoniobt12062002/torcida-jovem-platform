"use client";

import { type ReactNode, useState } from "react";

import { MoneyField } from "@/components/app/money-field";
import { QuantityField } from "@/components/app/quantity-field";
import { Checkbox } from "@/components/ui/checkbox";
import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { NativeSelect } from "@/components/ui/native-select";
import { Separator } from "@/components/ui/separator";
import { Textarea } from "@/components/ui/textarea";

import { SpecimenBlock, SpecimenCaption, SpecimenSection } from "./section";

// Campos (DSC-03, DSP-04): Field com cada controle em repouso, inválido com
// mensagem, desabilitado, somente leitura e opcional, quando o estado existe.
// O foco visível aparece ao navegar com Tab.

const GRID = "grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3";

function State({ name, children }: { name: string; children: ReactNode }) {
  return (
    <div className="flex flex-col gap-2">
      <SpecimenCaption>{name}</SpecimenCaption>
      {children}
    </div>
  );
}

function ContaOptions() {
  return (
    <>
      <option value="">Selecione a conta</option>
      <option value="associacoes">Associações</option>
      <option value="loja">Loja</option>
      <option value="eventos">Eventos</option>
    </>
  );
}

function NumericFields() {
  const [valor, setValor] = useState<number | null>(12550);
  const [valorInvalido, setValorInvalido] = useState<number | null>(null);
  const [quantidade, setQuantidade] = useState<number | null>(3);
  const [ajuste, setAjuste] = useState<number | null>(null);
  return (
    <>
      <SpecimenBlock title="MoneyField">
        <div className={GRID}>
          <State name="Repouso">
            <MoneyField label="Valor" value={valor} onChange={setValor} description="Em reais, com vírgula nos centavos." />
          </State>
          <State name="Inválido">
            <MoneyField label="Valor" value={valorInvalido} onChange={setValorInvalido} error="Informe o valor." />
          </State>
          <State name="Desabilitado">
            <MoneyField label="Valor" value={50000} onChange={() => {}} disabled />
          </State>
        </div>
      </SpecimenBlock>
      <SpecimenBlock title="QuantityField">
        <div className={GRID}>
          <State name="Repouso (positive)">
            <QuantityField label="Quantidade" mode="positive" value={quantidade} onChange={setQuantidade} />
          </State>
          <State name="Inválido (nonZero)">
            <QuantityField
              label="Ajuste de estoque"
              mode="nonZero"
              value={ajuste}
              onChange={setAjuste}
              error="Informe um número diferente de zero."
              description="Use - para saída."
            />
          </State>
          <State name="Desabilitado">
            <QuantityField label="Quantidade" mode="positive" value={10} onChange={() => {}} disabled />
          </State>
        </div>
        <SpecimenCaption>
          MoneyField e QuantityField não têm somente leitura nem marca de opcional: os estados que existem estão acima.
        </SpecimenCaption>
      </SpecimenBlock>
    </>
  );
}

export function FieldsSection() {
  return (
    <SpecimenSection
      id="campos"
      title="Campos"
      description="Rótulo sempre visível acima do campo; opcional marcado no rótulo; erro no campo, com ícone e texto. Use Tab para ver o foco."
    >
      <SpecimenBlock title="Input">
        <div className={GRID}>
          <State name="Repouso">
            <Field>
              <FieldLabel>Descrição</FieldLabel>
              <Input placeholder="Ex.: mensalidades de março…" />
              <FieldDescription>Aparece na lista de lançamentos.</FieldDescription>
            </Field>
          </State>
          <State name="Inválido com mensagem">
            <Field invalid>
              <FieldLabel>E-mail</FieldLabel>
              <Input type="email" defaultValue="pessoa@exemplo" spellCheck={false} autoCapitalize="none" />
              <FieldError>Informe um e-mail válido.</FieldError>
            </Field>
          </State>
          <State name="Desabilitado">
            <Field disabled>
              <FieldLabel>Conta</FieldLabel>
              <Input defaultValue="Associações" disabled />
            </Field>
          </State>
          <State name="Somente leitura">
            <Field>
              <FieldLabel>Código</FieldLabel>
              <Input defaultValue="fb6e3d8f" readOnly />
            </Field>
          </State>
          <State name="Opcional">
            <Field>
              <FieldLabel optional>Observação</FieldLabel>
              <Input />
            </Field>
          </State>
        </div>
      </SpecimenBlock>

      <SpecimenBlock title="Textarea">
        <div className={GRID}>
          <State name="Repouso">
            <Field>
              <FieldLabel>Motivo</FieldLabel>
              <Textarea placeholder="Explique o motivo do cancelamento…" />
            </Field>
          </State>
          <State name="Inválido com mensagem">
            <Field invalid>
              <FieldLabel>Motivo</FieldLabel>
              <Textarea />
              <FieldError>Informe o motivo.</FieldError>
            </Field>
          </State>
          <State name="Desabilitado">
            <Field disabled>
              <FieldLabel>Motivo</FieldLabel>
              <Textarea defaultValue="Lançamento duplicado." disabled />
            </Field>
          </State>
          <State name="Somente leitura">
            <Field>
              <FieldLabel>Histórico</FieldLabel>
              <Textarea defaultValue="Criado pela Tesouraria em 02/03/2026." readOnly />
            </Field>
          </State>
          <State name="Opcional">
            <Field>
              <FieldLabel optional>Observação</FieldLabel>
              <Textarea />
            </Field>
          </State>
        </div>
      </SpecimenBlock>

      <SpecimenBlock title="NativeSelect">
        <div className={GRID}>
          <State name="Repouso">
            <Field>
              <FieldLabel>Conta</FieldLabel>
              <NativeSelect defaultValue="">
                <ContaOptions />
              </NativeSelect>
            </Field>
          </State>
          <State name="Inválido com mensagem">
            <Field invalid>
              <FieldLabel>Conta</FieldLabel>
              <NativeSelect defaultValue="">
                <ContaOptions />
              </NativeSelect>
              <FieldError>Escolha a conta.</FieldError>
            </Field>
          </State>
          <State name="Desabilitado">
            <Field disabled>
              <FieldLabel>Conta</FieldLabel>
              <NativeSelect defaultValue="loja" disabled>
                <ContaOptions />
              </NativeSelect>
            </Field>
          </State>
          <State name="Opcional">
            <Field>
              <FieldLabel optional>Conta de destino</FieldLabel>
              <NativeSelect defaultValue="">
                <ContaOptions />
              </NativeSelect>
            </Field>
          </State>
        </div>
        <SpecimenCaption>O select nativo não tem somente leitura no HTML.</SpecimenCaption>
      </SpecimenBlock>

      <SpecimenBlock title="Checkbox">
        <div className={GRID}>
          <State name="Repouso">
            <Field>
              <FieldLabel>
                <Checkbox />
                Lembrar neste aparelho
              </FieldLabel>
            </Field>
          </State>
          <State name="Marcado">
            <Field>
              <FieldLabel>
                <Checkbox defaultChecked />
                Enviar comprovante por e-mail
              </FieldLabel>
            </Field>
          </State>
          <State name="Inválido com mensagem">
            <Field invalid>
              <FieldLabel>
                <Checkbox />
                Confirmo que os valores conferem
              </FieldLabel>
              <FieldError>É preciso confirmar.</FieldError>
            </Field>
          </State>
          <State name="Desabilitado">
            <Field disabled>
              <FieldLabel>
                <Checkbox disabled defaultChecked />
                Lançamento conferido
              </FieldLabel>
            </Field>
          </State>
        </div>
      </SpecimenBlock>

      <Separator />

      <NumericFields />
    </SpecimenSection>
  );
}
