# .specs

Fonte oficial das especificações da TJ Platform, no fluxo **Spec Driven Development** com a skill `tlc-spec-driven`:

```
Specify → Design → Tasks → Implement → Validate
```

```
.specs/
├── STATE.md            decisões do projeto (AD-NNN) e snapshot de retomada
└── features/
    └── <modulo>-<recorte>/
        ├── spec.md         requisitos com IDs e critérios de aceite em EARS
        ├── context.md      decisões do usuário em pontos ambíguos (quando houver)
        ├── design.md       arquitetura e componentes (features grandes)
        ├── tasks.md        tarefas atômicas com verificação (features grandes)
        └── validation.md   relatório do verificador (PASS/FAIL com evidências)
```

- Artefatos são criados sob demanda: um arquivo só existe quando a fase correspondente produziu conteúdo.
- Decisões que valem para todas as features entram em `STATE.md`; decisões locais ficam no `design.md` da feature.
- Decisões arquiteturais completas ficam em [`docs/adr/`](../docs/adr/README.md).
- Os documentos de produto em `docs/` serão convertidos gradualmente para specs neste formato.
