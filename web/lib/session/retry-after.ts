// FND-03 AC3: texto em português para o tempo de espera do cabeçalho
// Retry-After, que pode vir em segundos ou como data HTTP (RFC 9110).

function plural(n: number, singular: string, pluralForm: string) {
  return `${n} ${n === 1 ? singular : pluralForm}`;
}

function secondsFrom(header: string, now: number): number | null {
  const value = header.trim();
  if (/^\d+$/.test(value)) return Number(value);
  // Data HTTP: tem de ter letras (evita que "1.5" vire uma data).
  if (!/[a-z]/i.test(value)) return null;
  const at = Date.parse(value);
  if (Number.isNaN(at)) return null;
  return Math.ceil((at - now) / 1000);
}

/** "45 segundos", "15 minutos", "1 hora e 30 minutos"; null sem tempo válido. */
export function formatRetryAfter(
  header: string | null | undefined,
  now: number = Date.now(),
): string | null {
  if (!header) return null;
  const seconds = secondsFrom(header, now);
  if (seconds === null || !Number.isFinite(seconds) || seconds <= 0) return null;

  if (seconds < 60) return plural(seconds, "segundo", "segundos");
  const totalMinutes = Math.ceil(seconds / 60);
  if (totalMinutes < 60) return plural(totalMinutes, "minuto", "minutos");
  const hours = Math.floor(totalMinutes / 60);
  const minutes = totalMinutes % 60;
  const hoursText = plural(hours, "hora", "horas");
  return minutes === 0 ? hoursText : `${hoursText} e ${plural(minutes, "minuto", "minutos")}`;
}
