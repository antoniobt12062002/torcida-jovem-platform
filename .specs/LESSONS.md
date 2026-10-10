# LESSONS - auto-maintained by scripts/lessons.py

> Machine-owned. Do NOT hand-edit. Changes are overwritten on the next `lessons.py` write.
> Canonical state lives in `.specs/lessons.json`. Edit lessons only via the script.
> promote_threshold=2 distinct features · window_days=45 · quarantine_threshold=2

## Confirmed (load these at Specify/Design)

Corroborated across multiple features. Safe to apply as guidance.

_none_

## Candidates (under observation - do NOT load as guidance yet)

Seen once or not yet corroborated. Tracked, not trusted.

### L-001 - For a service that writes to an external store before a DB transaction, add a test that injects a failing store double and asserts zero rows were written at the service level, not just that the adapter itself returns an error.
- signal: `ac_gap` · recurrence: 1 feature(s) · scope: `platform/documents` · harmful: 0
- features: fundacao-documentos
- evidence: spec.md Edge Cases ("storage unreachable during a store operation"); api/internal/platform/documents/service.go:220-224 (platform/documents)
- last seen: 2026-09-27T18:13:59Z

### L-002 - Rodar o gate de auditoria de dependencias no inicio de cada rodada de integracao e da verificacao, porque um aviso novo reprova uma dependencia transitiva ja presente no lockfile sem nenhuma mudanca de codigo
- signal: `gate_fail` · recurrence: 1 feature(s) · scope: `web/gates` · harmful: 0
- features: web
- evidence: validation.md Gate Check: pnpm audit --audit-level high EXIT=1 (GHSA-68fv-2mgg-jv7q, web/pnpm-lock.yaml:3307) (web/gates)
- last seen: 2026-10-08T16:21:07Z

### L-003 - Quando o criterio proibe guardar um valor sensivel ou efemero em cache, o teste deve inspecionar o QueryClient (consultas e mutacoes) em busca do valor, nao so contar requisicoes
- signal: `surviving_mutant` · recurrence: 1 feature(s) · scope: `web/tests` · harmful: 0
- features: web
- evidence: mutante O06; FWB-05 AC5; web/features/financeiro/comprovantes/comprovantes-panel.test.tsx:235 (web/tests)
- last seen: 2026-10-08T16:21:07Z

### L-004 - Criterio de layout responsivo deve citar o ponto de quebra exato e dizer como ele e verificado, porque jsdom nao avalia media queries
- signal: `spec_precision_gap` · recurrence: 1 feature(s) · scope: `web/spec` · harmful: 0
- features: web
- evidence: SP-1; FND-05 AC2; web/components/app/app-shell.test.tsx:87 (web/spec)
- last seen: 2026-10-08T16:21:08Z

### L-005 - Mensagem derivada de Retry-After deve ter o formato do tempo de espera definido na spec
- signal: `spec_precision_gap` · recurrence: 1 feature(s) · scope: `web/spec` · harmful: 0
- features: web
- evidence: SP-2; FND-03 AC3 e ACS-01 AC6; web/app/(public)/entrar/login-form.test.tsx:104 (web/spec)
- last seen: 2026-10-08T16:21:08Z

### L-006 - Criterio que exige motivo deve declarar o tamanho minimo que a API impoe naquele endpoint
- signal: `spec_precision_gap` · recurrence: 1 feature(s) · scope: `web/spec` · harmful: 0
- features: web
- evidence: SP-3; USR-04 AC4; web/features/identity/usuarios/dialogs/admin-dialogs.test.tsx:248 (web/spec)
- last seen: 2026-10-08T16:21:08Z

### L-007 - Criterio que pede confirmacao deve dizer se ela e um passo separado ou o proprio dialogo do formulario
- signal: `spec_precision_gap` · recurrence: 1 feature(s) · scope: `web/spec` · harmful: 0
- features: web
- evidence: SP-4; EWB-03 AC3; web/features/estoque/ajustes/ajuste-dialog.test.tsx:58 (web/spec)
- last seen: 2026-10-08T16:21:08Z

### L-008 - Criterio sobre cabecalhos preservados por proxy ou rewrite deve apontar o item de smoke que confere cada cabecalho, porque o teste da configuracao nao os exercita
- signal: `spec_precision_gap` · recurrence: 1 feature(s) · scope: `web/spec` · harmful: 0
- features: web
- evidence: SP-5; FND-01 AC1; web/next.config.test.ts:13 (web/spec)
- last seen: 2026-10-08T16:21:10Z

### L-009 - Criterio de registro de evidencia deve dizer se a data e por item ou por execucao
- signal: `spec_precision_gap` · recurrence: 1 feature(s) · scope: `web/spec` · harmful: 0
- features: web
- evidence: SP-6; INT-03 AC1; .specs/features/web/smoke.md:3 (web/spec)
- last seen: 2026-10-08T16:21:10Z

### L-010 - Teste de visibilidade por permissao isola a permissao da acao: um caso so com ela e um caso com as outras escritas do modulo sem ela
- signal: `surviving_mutant` · recurrence: 1 feature(s) · scope: `web,tests,permissions` · harmful: 0
- features: web
- evidence: O02-O09 (validation.md, ciclo 1) (web,tests,permissions)
- last seen: 2026-10-08T22:28:11Z

### L-011 - Fixture de lista cuja ordem a spec define usa valores distintos e fora de ordem na chave de ordenacao provavel
- signal: `surviving_mutant` · recurrence: 1 feature(s) · scope: `web,tests,fixtures` · harmful: 0
- features: web
- evidence: O01 (validation.md, ciclo 1) (web,tests,fixtures)
- last seen: 2026-10-08T22:28:11Z

### L-012 - Quando um AC nomeia mais de um code da API para o mesmo comportamento, parametrize o teste por todos os codes nomeados
- signal: `surviving_mutant` · recurrence: 1 feature(s) · scope: `web,tests,session` · harmful: 0
- features: web
- evidence: S06 (web/lib/api/client.ts:81) / FND-03 AC6 (web,tests,session)
- last seen: 2026-10-09T03:53:57Z

### L-013 - Trava ou flag que impede acao repetida precisa de teste que repete a acao depois da condicao de liberacao, nao so de teste do bloqueio
- signal: `surviving_mutant` · recurrence: 1 feature(s) · scope: `web,tests,session` · harmful: 0
- features: web
- evidence: V07 (web/lib/session/session-provider.tsx:106) / FND-03 AC6 (web,tests,session)
- last seen: 2026-10-09T22:06:16Z

## Quarantined (failed when applied - ignore)

A confirmed lesson that recurred alongside failure. Kept for the maintainer to review.

_none_
