// Formatação de datas e autoria para as telas (FND-05 AC7, WEB-D-014). Valores
// em dinheiro usam formatBRL de lib/money.

const TIME_ZONE = "America/Sao_Paulo";
const DATE_ONLY = /^(\d{4})-(\d{2})-(\d{2})$/;

const dateTimeParts = new Intl.DateTimeFormat("pt-BR", {
  timeZone: TIME_ZONE,
  day: "2-digit",
  month: "2-digit",
  year: "numeric",
  hour: "2-digit",
  minute: "2-digit",
  hourCycle: "h23",
});

function parts(value: string | Date): Record<string, string> {
  const date = typeof value === "string" ? new Date(value) : value;
  if (Number.isNaN(date.getTime())) throw new RangeError(`data inválida: ${String(value)}`);
  const out: Record<string, string> = {};
  for (const part of dateTimeParts.formatToParts(date)) out[part.type] = part.value;
  return out;
}

/**
 * Data em pt-BR ("04/10/2026"). Um instante é lido no fuso de São Paulo; uma
 * data de calendário ("2026-10-04") é mostrada como está, sem fuso.
 */
export function formatDate(value: string | Date): string {
  if (typeof value === "string") {
    const match = DATE_ONLY.exec(value);
    if (match) return `${match[3]}/${match[2]}/${match[1]}`;
  }
  const p = parts(value);
  return `${p.day}/${p.month}/${p.year}`;
}

/** Data e hora em pt-BR no fuso de São Paulo ("04/10/2026 09:00"). */
export function formatDateTime(value: string | Date): string {
  const p = parts(value);
  return `${p.day}/${p.month}/${p.year} ${p.hour}:${p.minute}`;
}

/** Identificador curto: os 8 primeiros caracteres do UUID. */
export function shortId(id: string): string {
  return id.slice(0, 8);
}

/** Autoria exibida: "você" para a pessoa da sessão, senão o identificador curto. */
export function authorLabel(id: string, sessionUserId: string | null | undefined): string {
  return sessionUserId != null && id === sessionUserId ? "você" : shortId(id);
}
