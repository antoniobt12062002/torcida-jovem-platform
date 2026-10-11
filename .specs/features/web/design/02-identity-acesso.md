# Identity — Acesso e Senha (Web) — Design

**Spec**: `.specs/features/web/spec/02-identity-acesso.md`
**Pré-requisito**: sub-spec 01 até F2 integrada (sessão, rotas, kit de UI).

## Architecture Overview

```
app/(app)/conta/senha/page.tsx ─┐
app/(public)/recuperar-acesso/ ─┼─→ features/identity/acesso/* (formulários + hooks)
app/(public)/redefinir-senha/  ─┘        └─→ lib/api/client.ts (identity) · lib/session
```

As páginas são finas: montam o formulário da feature. Toda lógica fica em `features/identity/acesso/`.

## Code Reuse Analysis

### Existing Components to Leverage

| Existente | Uso |
| --- | --- |
| `lib/session` (`useSession`, `routes.ts`, `safeNext`, formatação de `Retry-After`) | destino depois da troca, dica de tamanho, bloqueio |
| `lib/api/client.ts`, `lib/api/problem.ts`, `lib/forms/` | chamadas e erros |
| `components/ui` (`field`, `input`, `button`, `alert`) | formulários |

### Integration Points

| Ponto | Como |
| --- | --- |
| Portão de F2 | Já leva `must_change_password` para `/conta/senha`; esta sub-spec cria a página. |
| `/entrar` (F2) | `app/(public)/entrar/` pertence a F2, que já cria o link "Esqueci minha senha" para `routes.recuperarAcesso` (FND-03 AC10). F3 não altera `/entrar`; ACS-02 AC3 é conferido pela INT no smoke, quando as duas rotas existem. |
| Shell (F5) | No modo obrigatório, quem esconde a navegação é o shell (FND-05 AC8), a partir do contexto de sessão. F3 não altera `components/app/` e não depende de F5. |

## Components

| Arquivo | Função |
| --- | --- |
| `features/identity/acesso/errors.ts` | Catálogo `code → mensagem` e `code → campo` dos erros de senha (ACS-01 AC5, ACS-03 AC4-5) |
| `features/identity/acesso/change-password-form.tsx` | Formulário de troca; modo obrigatório e voluntário |
| `features/identity/acesso/use-change-password.ts` | Mutação `POST /auth/password`; em sucesso, invalida `queryKeys.session.me` |
| `features/identity/acesso/request-reset-form.tsx` | Formulário de pedido |
| `features/identity/acesso/reset-password-form.tsx` | Leitura do fragmento, `history.replaceState`, formulário de confirmação |
| `features/identity/acesso/read-reset-token.ts` | Função pura: fragmento → token ou nulo |
| `app/(app)/conta/senha/page.tsx` | Página de troca |
| `app/(public)/recuperar-acesso/page.tsx` | Página de pedido |
| `app/(public)/redefinir-senha/page.tsx` | Página de redefinição |

`/conta/senha` fica dentro de `app/(app)/` porque exige sessão. No modo obrigatório, o portão de F2 impede outras rotas (FND-03 AC4) e o shell de F5 esconde a navegação (FND-05 AC8); a página só mostra o formulário e o aviso.

## Data Models

`ChangePasswordRequest`, `PasswordResetRequest`, `PasswordResetConfirm` e `Accepted` dos `.d.ts` gerados.

## Error Handling Strategy

| `code` | Campo | Mensagem |
| --- | --- | --- |
| `invalid_current_password` | senha atual | "A senha atual está incorreta." |
| `password_unchanged` | nova senha | "A nova senha precisa ser diferente da atual." |
| `password_too_short` | nova senha | "A senha é curta demais." + dica de mínimo |
| `password_too_long` | nova senha | "A senha é longa demais." |
| `password_compromised` | nova senha | "Esta senha aparece em vazamentos conhecidos. Escolha outra." |
| `invalid_reset_token` | formulário | "O link expirou ou já foi usado." + link para pedir outro |
| `password_change_blocked` (429) | formulário | "Muitas tentativas. Tente de novo em …" |

## Risks & Concerns

| Risco | Mitigação |
| --- | --- |
| Token vazar por `Referer` ou histórico | Fragmento nunca vai ao servidor; removido com `replaceState` antes de qualquer navegação. |
| Hidratação ler `window.location.hash` no servidor | A leitura acontece em efeito do Client Component. |

## Tech Decisions

| Decisão | Motivo |
| --- | --- |
| Confirmação de senha só no front | A API recebe uma senha; a confirmação evita erro de digitação sem regra nova. |
