import type { ErrorCatalog } from "@/lib/api/problem";

import { formatBytes, MAX_UPLOAD_BYTES } from "./limits";

// Mensagens dos erros de comprovantes (FWB-05 AC3, AC4 e AC6).

export const MESSAGE_ARQUIVO_ACIMA_DO_LIMITE = `O arquivo excede o limite permitido de ${formatBytes(MAX_UPLOAD_BYTES)}. Escolha um arquivo menor.`;
export const MESSAGE_API_ACIMA_DO_LIMITE = "O arquivo excede o limite permitido. Escolha um arquivo menor.";
export const MESSAGE_EXTENSAO = "Extensão de arquivo não permitida. Use PDF, JPG, JPEG, PNG ou WEBP.";
export const MESSAGE_TIPO = "Tipo de arquivo não permitido. Use PDF, JPG, JPEG, PNG ou WEBP.";
export const MESSAGE_TIPO_DIVERGENTE = "O conteúdo do arquivo não corresponde à extensão do nome.";
export const MESSAGE_DOCUMENTO_NAO_ENCONTRADO = "Comprovante não encontrado.";
export const MESSAGE_LANCAMENTO_NAO_ENCONTRADO = "Lançamento não encontrado.";

export const comprovantesCatalog: ErrorCatalog = {
  payload_too_large: MESSAGE_API_ACIMA_DO_LIMITE,
  document_too_large: MESSAGE_API_ACIMA_DO_LIMITE,
  document_extension_not_allowed: MESSAGE_EXTENSAO,
  document_type_not_allowed: MESSAGE_TIPO,
  document_type_mismatch: MESSAGE_TIPO_DIVERGENTE,
  document_not_found: MESSAGE_DOCUMENTO_NAO_ENCONTRADO,
  lancamento_nao_encontrado: MESSAGE_LANCAMENTO_NAO_ENCONTRADO,
};
