/**
 * Converte um valor de tela de e para o texto da query string.
 * `parse` recebe `null` quando o parâmetro não está na URL e devolve o padrão
 * para valor ausente ou inválido, sem lançar erro. `serialize` devolve `null`
 * para o padrão, o que remove o parâmetro da URL.
 */
export type Codec<T> = {
  parse: (raw: string | null) => T;
  serialize: (value: T) => string | null;
};

const ISO_DATE = /^(\d{4})-(\d{2})-(\d{2})$/;

/** `AAAA-MM-DD` de um dia que existe no calendário (30 de fevereiro não passa). */
function isIsoDate(value: string) {
  const match = ISO_DATE.exec(value);
  if (!match) return false;
  const [year, month, day] = match.slice(1).map(Number);
  const date = new Date(Date.UTC(year, month - 1, day));
  return date.getUTCFullYear() === year && date.getUTCMonth() === month - 1 && date.getUTCDate() === day;
}

export const codecs = {
  /** Texto livre (busca). Padrão: vazio. */
  string: (fallback = ""): Codec<string> => ({
    parse: (raw) => raw ?? fallback,
    serialize: (value) => (value === fallback ? null : value),
  }),

  /** Um dos valores da lista (filtro de seleção, ordenação, aba). Fora da lista, vale o padrão. */
  oneOf: <T extends string>(values: readonly T[], fallback: T | ""): Codec<T | ""> => {
    const isValue = (value: string | null): value is T => values.includes(value as T);
    return {
      parse: (raw) => (isValue(raw) ? raw : fallback),
      serialize: (value) => (value !== fallback && isValue(value) ? value : null),
    };
  },

  /** Data `AAAA-MM-DD`. Padrão: vazio; data inválida cai no padrão. */
  isoDate: (): Codec<string> => ({
    parse: (raw) => (raw !== null && isIsoDate(raw) ? raw : ""),
    serialize: (value) => (isIsoDate(value) ? value : null),
  }),
};
