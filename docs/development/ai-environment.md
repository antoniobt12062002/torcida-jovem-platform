# Ambiente de desenvolvimento com IA

As skills de IA são ferramentas do ambiente de cada desenvolvedor. **Não são versionadas** neste repositório (as pastas `.claude/`, `.cursor/`, `.windsurf/` e `.agents/` estão no `.gitignore`). Este documento registra quais são usadas, para quê e quando acioná-las, para que o ambiente seja reproduzível.

## Skills utilizadas

| Skill | Finalidade | Quando acionar |
|---|---|---|
| `tlc-spec-driven` | Fluxo Spec Driven: Specify, Design, Tasks, Implement, Validate, com validadores e verificador independente | Ao iniciar qualquer funcionalidade; mantém `.specs/` |
| `technical-design-doc-creator` | Design docs técnicos completos, por descoberta interativa | Temas transversais ou complexos (autenticação, auditoria, integração de pagamento) antes de implementar |
| `create-adr` | Registro de decisões de arquitetura no formato MADR em `docs/adr/` | Sempre que uma decisão relevante for tomada; nunca para decisões ainda em aberto |
| `security-best-practices` | Revisão de segurança para Go e JavaScript/TypeScript | No design e antes de abrir PR que toque autenticação, sessões, dinheiro, uploads ou entrada de usuário |
| `the-judge` | Revisão de PR baseada em evidências, publicada como review no GitHub | Antes de mesclar qualquer PR relevante |
| `spec-driven-eval` | Avaliação da qualidade do processo Spec Driven | Somente por invocação explícita do mantenedor (`/spec-driven-eval`); o agente não a executa por conta própria |
| `vercel-deploy` | Deploy na Vercel | Não usada agora; só se a hospedagem do front for a Vercel |

O plugin `superpowers` (brainstorming, planos, execução) foi usado nas fases iniciais; os artefatos resultantes estão arquivados em [`docs/archive/superpowers/`](../archive/superpowers/plans/2026-09-26-github-cicd.md). A partir da governança, o fluxo oficial é o do `tlc-spec-driven`.

## Instalação

```bash
npx @tech-leads-club/agent-skills install --skill technical-design-doc-creator
npx @tech-leads-club/agent-skills install --skill the-judge
npx @tech-leads-club/agent-skills install --skill spec-driven-eval
npx @tech-leads-club/agent-skills install --skill create-adr
```

`tlc-spec-driven`, `security-best-practices` e `vercel-deploy` já constam no lockfile local do ambiente (`.agents/.skill-lock.json`); instale-as da mesma forma, pelo mesmo pacote, se estiverem ausentes.

## Regras de uso

- Decisões arquiteturais viram ADR; decisões de projeto que valem para todas as features entram também em `.specs/STATE.md`.
- Nenhuma skill grava segredos ou dados pessoais no repositório (ele é público).
- Ações remotas (push, PR, merge, deploy) exigem autorização explícita do mantenedor; a aprovação de uma spec autoriza apenas implementação e commits locais.
- O `the-judge` publica a revisão com a conta do mantenedor; como o GitHub não permite aprovar o próprio PR, o resultado funciona como comentário, e não como aprovação formal.
- Instruções para os agentes ficam no [`CLAUDE.md`](../../CLAUDE.md) da raiz.
