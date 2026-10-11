// Limite efetivo de upload na camada Web (FWB-05 AC3, WEB-D-009, correção de
// 2026-10-04): 10 MiB − 64 KiB, margem para o envelope multipart/form-data,
// porque o rewrite do Next trunca corpos acima de cerca de 10 MiB. É mitigação
// do front, não regra de negócio: o limite da API continua 10 MiB e é a
// autoridade. Constante única; nenhum outro ponto repete o número.
export const MAX_UPLOAD_BYTES = 10_420_224;

/** Extensões sugeridas no seletor (lista de platform/documents); a API decide. */
export const ACCEPTED_EXTENSIONS = ".pdf,.jpg,.jpeg,.png,.webp";

const oneDecimal = new Intl.NumberFormat("pt-BR", { minimumFractionDigits: 1, maximumFractionDigits: 1 });

/** Tamanho legível em pt-BR, em base 1024 ("512 bytes", "1,0 KB", "9,9 MB"). */
export function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} bytes`;
  if (bytes < 1024 * 1024) return `${oneDecimal.format(Math.floor((bytes / 1024) * 10) / 10)} KB`;
  return `${oneDecimal.format(Math.floor((bytes / (1024 * 1024)) * 10) / 10)} MB`;
}

/** O arquivo pode ser enviado? Até MAX_UPLOAD_BYTES, inclusive. */
export function withinUploadLimit(size: number): boolean {
  return size <= MAX_UPLOAD_BYTES;
}
