import type { Conta, Lancamento, StatusLancamento, TipoConta } from "./labels";

// Filtro e ordenação dos lançamentos no cliente (FWB-03 AC1, WEB-D-016) e
// escolha das contas oferecidas num lançamento. Funções puras.

export type LancamentoFilters = {
  tipo: TipoConta | "";
  status: StatusLancamento | "";
  contaId: string;
  /** Data de criação inicial, inclusiva ("AAAA-MM-DD"). */
  de: string;
  /** Data de criação final, inclusiva ("AAAA-MM-DD"). */
  ate: string;
};

export const EMPTY_FILTERS: LancamentoFilters = { tipo: "", status: "", contaId: "", de: "", ate: "" };

const calendarDate = new Intl.DateTimeFormat("en-CA", {
  timeZone: "America/Sao_Paulo",
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
});

/** Data de calendário ("AAAA-MM-DD") de um instante, no fuso de São Paulo. */
export function localDate(iso: string): string {
  return calendarDate.format(new Date(iso));
}

/** Aplica os filtros e ordena do mais novo para o mais antigo (cópia). */
export function filterLancamentos(
  items: readonly Lancamento[],
  filters: LancamentoFilters,
): Lancamento[] {
  return items
    .filter((l) => {
      if (filters.tipo && l.tipo !== filters.tipo) return false;
      if (filters.status && l.status !== filters.status) return false;
      if (filters.contaId && l.conta_id !== filters.contaId) return false;
      if (filters.de || filters.ate) {
        const day = localDate(l.criado_em);
        if (filters.de && day < filters.de) return false;
        if (filters.ate && day > filters.ate) return false;
      }
      return true;
    })
    .sort((a, b) => Date.parse(b.criado_em) - Date.parse(a.criado_em));
}

/** Contas que um lançamento do tipo dado pode usar: ativas e do mesmo tipo, por nome. */
export function contasParaLancamento(items: readonly Conta[] | undefined, tipo: TipoConta): Conta[] {
  return (items ?? [])
    .filter((c) => c.ativo && c.tipo === tipo)
    .sort((a, b) => a.nome.localeCompare(b.nome, "pt-BR"));
}
