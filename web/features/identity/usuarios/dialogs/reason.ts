// Motivo mínimo conferido no front para a promoção e a senha temporária
// (USR-04 AC2, USR-05 AC1). A API continua conferindo (reason_required).

export const MIN_REASON_LENGTH = 10;

export const REASON_HINT = `Mínimo de ${MIN_REASON_LENGTH} caracteres.`;

/** Motivo aparado com ao menos 10 caracteres úteis. */
export function isReasonValid(reason: string): boolean {
  return reason.trim().length >= MIN_REASON_LENGTH;
}
