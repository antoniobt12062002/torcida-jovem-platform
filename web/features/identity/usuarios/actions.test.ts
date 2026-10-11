import { describe, expect, it } from "vitest";

import { ids, type User, user } from "@/test/msw/fixtures";

import { ACTION_LABELS, availableActions, type UserAction } from "./actions";

// USR-03, USR-04 AC1/AC3/AC4/AC6, USR-05 e o caso de borda da própria linha:
// cada ação de linha depende só da permissão nomeada na spec e do estado do
// usuário (FND-04 AC3).

const SESSION_ID = ids.user;

const MEMBERSHIP = { reason: "Assumiu a tesouraria", granted_at: "2026-10-01T12:00:00Z" };

const semVinculoAtivo = user({ active: true, admin_membership: null });
const semVinculoInativo = user({ active: false, admin_membership: null });
const comVinculoAtivo = user({
  active: true,
  roles: ["ASSOCIADO", "TESOURARIA"],
  admin_membership: MEMBERSHIP,
});
const comVinculoInativo = user({
  active: false,
  roles: ["ASSOCIADO", "TESOURARIA"],
  admin_membership: MEMBERSHIP,
});

const STATES: ReadonlyArray<{ label: string; target: User }> = [
  { label: "sem vínculo, ativo", target: semVinculoAtivo },
  { label: "sem vínculo, inativo", target: semVinculoInativo },
  { label: "com vínculo, ativo", target: comVinculoAtivo },
  { label: "com vínculo, inativo", target: comVinculoInativo },
];

/** Permissões que liberam ações de linha, na spec. */
const ROW_PERMISSIONS = [
  "identity:admin:grant",
  "identity:role:assign",
  "identity:admin:revoke",
  "identity:user:reset_password",
  "identity:user:update",
] as const;

type RowPermission = (typeof ROW_PERMISSIONS)[number];

/** Permissões do módulo que não liberam nenhuma ação de linha. */
const NEUTRAL = ["identity:user:read", "identity:user:create"] as const;

/**
 * Tabela da spec, escrita por extenso: para cada permissão isolada, a ação
 * oferecida em cada estado (na ordem de STATES).
 */
const ALONE: Readonly<Record<RowPermission, readonly UserAction[][]>> = {
  "identity:admin:grant": [["grant"], ["grant"], [], []],
  "identity:role:assign": [[], [], ["setRoles"], ["setRoles"]],
  "identity:admin:revoke": [[], [], ["revoke"], ["revoke"]],
  "identity:user:reset_password": [
    ["resetPassword"],
    ["resetPassword"],
    ["resetPassword"],
    ["resetPassword"],
  ],
  "identity:user:update": [["deactivate"], ["reactivate"], ["deactivate"], ["reactivate"]],
};

/** Com todas as permissões, por estado (na ordem de STATES). */
const ALL: readonly UserAction[][] = [
  ["grant", "resetPassword", "deactivate"],
  ["grant", "resetPassword", "reactivate"],
  ["setRoles", "revoke", "resetPassword", "deactivate"],
  ["setRoles", "revoke", "resetPassword", "reactivate"],
];

function subsets<T>(values: readonly T[]): T[][] {
  return values.reduce<T[][]>((acc, value) => [...acc, ...acc.map((set) => [...set, value])], [[]]);
}

describe("availableActions", () => {
  describe.each(ROW_PERMISSIONS.map((permission) => ({ permission })))(
    "$permission",
    ({ permission }) => {
      it.each(STATES.map((state, index) => ({ ...state, index })))(
        "sozinha ($label): libera só a sua ação",
        ({ target, index }) => {
          expect(availableActions(target, SESSION_ID, [...NEUTRAL, permission])).toEqual(
            ALONE[permission][index],
          );
        },
      );

      it.each(STATES.map((state, index) => ({ ...state, index })))(
        "ausente, com todas as outras ($label): a sua ação some e as demais ficam",
        ({ target, index }) => {
          const others = ROW_PERMISSIONS.filter((other) => other !== permission);
          const removed = ALONE[permission][index];
          expect(availableActions(target, SESSION_ID, [...NEUTRAL, ...others])).toEqual(
            ALL[index].filter((action) => !removed.includes(action)),
          );
        },
      );
    },
  );

  it.each(STATES.map((state, index) => ({ ...state, index })))(
    "com todas as permissões ($label): só as ações compatíveis com o estado, na ordem do menu",
    ({ target, index }) => {
      expect(availableActions(target, SESSION_ID, [...NEUTRAL, ...ROW_PERMISSIONS])).toEqual(
        ALL[index],
      );
    },
  );

  it("toda combinação de permissões × estado é a união das ações de cada permissão", () => {
    let checked = 0;
    for (const granted of subsets(ROW_PERMISSIONS)) {
      STATES.forEach(({ target }, index) => {
        const expected = ALL[index].filter((action) =>
          granted.some((permission) => ALONE[permission][index].includes(action)),
        );
        expect(
          availableActions(target, SESSION_ID, [...NEUTRAL, ...granted]),
          `${granted.join(" + ") || "nenhuma"} / ${STATES[index].label}`,
        ).toEqual(expected);
        checked += 1;
      });
    }
    expect(checked).toBe(32 * 4);
  });

  it.each(STATES)(
    "só leitura e criação ($label): nenhuma ação de linha",
    ({ target }) => {
      expect(availableActions(target, SESSION_ID, [...NEUTRAL])).toEqual([]);
      expect(availableActions(target, SESSION_ID, [])).toEqual([]);
    },
  );

  it("permissões de outros módulos ou parecidas não liberam nada", () => {
    const foreign = [
      "financeiro:conta:read",
      "financeiro:lancamento:create",
      "estoque:produto:update",
      "identity:user:read",
      "identity:user:create",
      "identity:role:read",
      "identity:admin:read",
      "user:update",
      "identity:user:update:extra",
    ];
    for (const { target } of STATES) {
      expect(availableActions(target, SESSION_ID, foreign)).toEqual([]);
    }
  });

  describe("própria linha", () => {
    it.each(STATES)("nenhuma ação, mesmo com todas as permissões ($label)", ({ target }) => {
      const self = { ...target, id: SESSION_ID };
      expect(availableActions(self, SESSION_ID, [...NEUTRAL, ...ROW_PERMISSIONS])).toEqual([]);
    });

    it.each(ROW_PERMISSIONS.map((permission) => ({ permission })))(
      "nenhuma ação com $permission isolada",
      ({ permission }) => {
        for (const { target } of STATES) {
          expect(availableActions({ ...target, id: SESSION_ID }, SESSION_ID, [permission])).toEqual(
            [],
          );
        }
      },
    );

    it("a linha de outra pessoa não é afetada pelo id da sessão", () => {
      expect(availableActions(comVinculoAtivo, SESSION_ID, [...ROW_PERMISSIONS])).toEqual(ALL[2]);
    });

    it("sem id de sessão conhecido, decide só por permissão e estado", () => {
      expect(availableActions(comVinculoAtivo, null, [...ROW_PERMISSIONS])).toEqual(ALL[2]);
      expect(availableActions(semVinculoInativo, undefined, [...ROW_PERMISSIONS])).toEqual(ALL[1]);
    });
  });

  it("rótulos das ações", () => {
    expect(ACTION_LABELS).toEqual({
      deactivate: "Desativar",
      reactivate: "Reativar",
      grant: "Promover a administrativo",
      setRoles: "Editar papéis",
      revoke: "Retirar acesso administrativo",
      resetPassword: "Gerar senha temporária",
    });
  });
});
