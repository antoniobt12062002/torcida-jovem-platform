import type { ErrorCatalog } from "@/lib/api/problem";

// Catálogo dos erros da administração de usuários (USR-02 a USR-05). A API
// devolve o `code` no corpo; o que não está aqui segue as mensagens genéricas
// da fundação (FND-06).

export const MESSAGE_EMAIL_TAKEN = "Já existe um usuário com este e-mail.";
export const MESSAGE_INVALID_NAME = "Informe um nome válido.";
export const MESSAGE_INVALID_EMAIL = "Informe um e-mail válido.";
export const MESSAGE_PASSWORD_TOO_SHORT = "A senha é curta demais. Mínimo de 8 caracteres.";
export const MESSAGE_PASSWORD_TOO_LONG = "A senha é longa demais.";
export const MESSAGE_PASSWORD_COMPROMISED =
  "Esta senha aparece em vazamentos conhecidos. Escolha outra.";

export const usersCatalog: ErrorCatalog = {
  // USR-02
  email_taken: MESSAGE_EMAIL_TAKEN,
  invalid_name: MESSAGE_INVALID_NAME,
  invalid_email: MESSAGE_INVALID_EMAIL,
  password_too_short: MESSAGE_PASSWORD_TOO_SHORT,
  password_too_long: MESSAGE_PASSWORD_TOO_LONG,
  password_compromised: MESSAGE_PASSWORD_COMPROMISED,
  // USR-03 a USR-05
  self_change_forbidden: "Você não pode fazer isso com a sua própria conta.",
  last_admin: "Não é possível: o sistema ficaria sem ninguém que possa conceder acesso.",
  privilege_escalation: "Você não pode conceder ou alterar permissões que não tem.",
  already_admin: "Esta pessoa já tem acesso administrativo.",
  not_admin: "Esta pessoa não tem acesso administrativo.",
  admin_membership_required:
    "Só é possível definir papéis administrativos para quem tem acesso administrativo.",
  user_inactive: "Esta pessoa está desativada. Reative a conta antes.",
  unknown_role: "Um dos papéis escolhidos não existe.",
  admin_role_required: "Escolha ao menos um papel administrativo.",
  reason_required: "Informe um motivo com pelo menos 10 caracteres.",
  not_found: "Usuário não encontrado.",
};
