# ADR-015: Superfícies web, rotas públicas e entrada de webhook do TJ Connect

- **Date**: 2026-10-13 (revisão 2, com as decisões M-1 a M-9 do mantenedor)
- **Status**: Proposed. Aprovado conceitualmente; ainda não aceito; não autoriza implementação.
- **Deciders**: @antoniobt12062002
- **Tags**: web, security, http, tj-connect

## Context and Problem Statement

Hoje a Web tem só o painel administrativo:

- `web/app/page.tsx` redireciona `/` para `/inicio`;
- o grupo `(app)` usa o shell administrativo;
- as telas são Client Components (WEB-D-002);
- a Web é *desktop-first* (WEB-D-005).

Na API, o [AD-013](../../.specs/STATE.md) fecha a lista pública em `/healthz`, login e as duas rotas de recuperação. A cadeia pública exige `Origin` e a autenticada exige sessão e CSRF. O limite de tentativas existe por e-mail no login, não por IP.

O TJ Connect precisa de:

- um site público com SEO;
- a área Meu TJ e a área do fiscal, ambas *mobile-first*;
- cadastro público com verificação de e-mail;
- um catálogo público de planos;
- um webhook chamado servidor a servidor pelo Mercado Pago, que não envia `Origin`, cookie nem CSRF.

O mantenedor aprovou as três experiências, a preservação do painel e o login unificado, com navegação pós-login baseada em permissões e vínculo associativo.

## Decision Drivers

- Negar por padrão continua valendo: cada rota pública é uma decisão explícita e testada.
- Uma identidade e uma sessão para todas as superfícies (TJ-ID-001); o cookie, `SameSite=Lax` e o rewrite same-origin ([ADR-010](010-acesso-do-navegador-a-api-por-rewrite-same-origin.md)) não mudam.
- O Design System ([AD-018](../../.specs/STATE.md)) é a base visual de todas as superfícies.
- A câmera do navegador (`getUserMedia`) só funciona em contexto seguro (HTTPS).

## Considered Options

1. **Quatro superfícies num único app Next.js** (site, Meu TJ, fiscal, painel), com cadeias HTTP novas e explícitas para o público e para o webhook.
2. **Apps Next.js separados** para o site público e para a plataforma.
3. **Site público estático fora da plataforma** (CMS ou gerador estático), com a plataforma só para áreas autenticadas.

## Decision Outcome

Chosen option: **1**, porque mantém uma sessão, um Design System e um pipeline, e isola o que muda numa cadeia própria por tipo de chamador.

### Web

| Superfície | Grupo de rotas | Renderização | Layout |
|---|---|---|---|
| Site público | `(site)` | Server Components com renderização estática ou revalidação, metadados e sitemap. Só leitura pública da API pelo servidor do Next | Responsivo, *mobile-first* |
| Meu TJ | `(meu-tj)` | Client Components, como o painel (WEB-D-002) | *Mobile-first* |
| Fiscal | `(fiscal)` | Client Components; câmera; somente online | *Mobile-first*, operação com uma mão, resultado em tela cheia |
| Painel | `(app)`, como hoje | Como hoje | Como hoje (WEB-D-005 continua valendo só aqui) |

- **`/` vira a home pública.** Isso substitui o critério INT-01 AC3 da feature `web` (redirecionar para `/inicio`) por uma alteração registrada. `/inicio` e as rotas atuais do painel não mudam.
- **Navegação pós-login**: o destino vem do contexto da sessão (permissões) e do vínculo associativo (consulta a `associados`). Quem tem permissão administrativa vai ao painel; quem não tem vai ao Meu TJ; o fiscal vê o acesso à área do fiscal. As três superfícies autenticadas oferecem troca entre si conforme o que a pessoa pode fazer. O portão no front é só navegação; a API continua sendo a autoridade.
- **Liberação comercial não bloqueia o público** (decisão M-5). O site, o catálogo público e o cadastro gratuito funcionam sem nenhuma liberação comercial. Um plano sem liberação vigente aparece como indisponível para contratação (o texto exato fica para a spec `10`). A **API** recusa iniciar o checkout de um plano sem liberação vigente, sem depender da interface.
- **Checkout Pro** (decisão M-2): o Meu TJ redireciona o comprador ao `checkout_url` do Mercado Pago. As páginas de retorno mostram só "pagamento em verificação" e consultam o estado na API; nunca exibem a adesão como quitada pelo retorno do provedor (AC-PAG-01).
- **Fiscal** (decisão M-7): a interface não oferece o próprio associado ao fiscal, mas a proibição de auto-check-in é verificada **no backend** ([ADR-016](016-permissoes-sensiveis-e-escopo-por-jogo.md)).
- **Leitor de QR**: a escolha (API nativa do navegador ou biblioteca) é dependência nova e fica para o Specify da área do fiscal, com verificação de suporte em iOS e Android. **Não verificado aqui.**

### API: cadeias e lista pública

- **Cadeia pública** (origem, limite de 64 KiB, validação do contrato) passa a incluir:
  - cadastro;
  - confirmação de e-mail;
  - reenvio de confirmação;
  - leitura do catálogo público da temporada (planos, preços e disponibilidade agregada de quota, sem dado pessoal).

  Toda rota pública nova entra numa lista explícita e testada.
- **Limite por IP e por identificador** em cadastro, login, recuperação, reenvio e webhook (TJ-ID-004, §16 Segurança). É a "limitação unificada" que já constava como decisão futura no `STATE.md` e passa a ser requisito.
- **Cadeia de webhook** (nova, só para provedores de pagamento):
  - limite de corpo pequeno;
  - **validação da assinatura do provedor antes de tudo** (HMAC do Mercado Pago, [ADR-012](012-cobrancas-e-eventos-persistentes.md));
  - validação do contrato;
  - sem `Origin`, sem sessão e sem CSRF;
  - limite de requisições;
  - resposta sem detalhes.

  Uma assinatura inválida é registrada e recusada.
- **A rota do fiscal** usa a cadeia autenticada normal: o check-in é uma escrita com sessão e CSRF, e a autorização por jogo fica no caso de uso ([ADR-016](016-permissoes-sensiveis-e-escopo-por-jogo.md)).

### Positive Consequences

- Uma conta e uma sessão em tudo; o presidente pode ser associado sem segundo cadastro.
- O site público tem SEO sem expor a API autenticada.
- O webhook fica isolado numa cadeia própria e verificável, sem afrouxar as demais.

### Negative Consequences

- O app Next.js passa a misturar páginas estáticas e de cliente; o build e os testes crescem.
- O AD-013 precisa ser emendado (lista pública maior e uma cadeia nova).
- A limitação por IP atrás do rewrite do Next exige confiar no cabeçalho de IP encaminhado só a partir do proxy conhecido. O Design precisa definir isso para não abrir falsificação.

## Pros and Cons of the Options

### 1. Quatro superfícies num app ✅ Chosen

- ✅ Uma sessão, um Design System, um deploy.
- ❌ App maior.

### 2. Apps separados

- ✅ Isolamento de build.
- ❌ Sessão entre origens ou subdomínios exige rever cookie, CORS e CSRF, e o [ADR-010](010-acesso-do-navegador-a-api-por-rewrite-same-origin.md) foi justamente para evitar isso.

### 3. Site estático externo

- ✅ SEO simples.
- ❌ Planos e disponibilidade de quota precisariam ser duplicados ou integrados à parte; outra ferramenta para manter.

## Links

- Emenda AD-013 e WEB-D-005 (só para as superfícies novas) e o critério INT-01 AC3 da feature `web`.
- [ADR-005](005-autenticacao-e-rbac.md), [ADR-010](010-acesso-do-navegador-a-api-por-rewrite-same-origin.md), [ADR-012](012-cobrancas-e-eventos-persistentes.md), [ADR-016](016-permissoes-sensiveis-e-escopo-por-jogo.md).
