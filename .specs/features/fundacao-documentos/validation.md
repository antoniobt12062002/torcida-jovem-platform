# Fundação Documentos Validation

**Date**: 2026-09-27
**Spec**: `.specs/features/fundacao-documentos/spec.md`
**Diff range**: `d29eb04..e98660f` (commit `e98660f`, squash-merged PR #38)
**Verifier**: independent sub-agent (author ≠ verifier)

---

## Task Completion

| Task | Status  | Notes |
| ---- | ------- | ----- |
| T1 (config S3 + URL TTL) | ✅ Done | `api/internal/config/config.go:213-260`, tested in `config_env_test.go` |
| T2 (interface `Storage`) | ✅ Done | `api/internal/platform/storage/storage.go` |
| T3 (emulador Garage + helper) | ✅ Done | `api/internal/platform/testutil/s3.go`, smoke test in `s3_test.go` |
| T4 (`S3Storage` real) | ✅ Done | `api/internal/platform/storage/s3.go` |
| T5 (migração `documents`) | ✅ Done | `api/migrations/000005_documents.{up,down}.sql`, tested in `documents_migration_test.go` |
| T6 (`Service.Store`/`AccessURL`) | ✅ Done | `api/internal/platform/documents/service.go` |
| T7 (`Service.ListByOwner`) | ✅ Done | `api/internal/platform/documents/list.go` |

All 7 tasks' "Done when" checklists in `tasks.md` are marked `[x]`; independently re-derived and confirmed against the diff (not taken on trust) — see AC table below.

---

## Spec-Anchored Acceptance Criteria

### DOC-01: Armazenamento de documentos

| Criterion | Spec-defined outcome | `file:line` + assertion | Result |
| --- | --- | --- | --- |
| AC1: store writes object + metadata row | key `documents/<uuid>`, all listed columns populated | `api/internal/platform/documents/service_test.go:183-214` `TestStoreWritesTheObjectAndTheMetadataRow` — asserts `doc.StorageKey == "documents/"+doc.ID`, row's `owner_type/owner_id/original_filename/content_type/size_bytes/storage_key/version/uploaded_by`, then downloads the object and compares bytes | ✅ PASS |
| AC2: SHA-256 over streamed bytes, hex | exact hex digest of the uploaded bytes | `service_test.go:217-240` `TestStoreComputesTheSHA256OverTheStreamedBytes` — computes `sha256.Sum256(content)` independently and asserts `doc.SHA256 == want` and `len(doc.SHA256) == 64` | ✅ PASS |
| AC3: >10 MiB rejected before writing metadata | error `document_too_large`, zero objects written | `service_test.go:243-262` `TestStoreRejectsAFileOver10MiB` — asserts `errors.Is(err, documents.ErrTooLarge)` and `objectCount` unchanged | ✅ PASS |
| AC4: content type not in allow-list rejected | error `document_type_not_allowed` | `service_test.go:265-277` `TestStoreRejectsAContentTypeNotInTheAllowList` (GIF bytes) — `errors.Is(err, documents.ErrTypeNotAllowed)` | ✅ PASS |
| AC5: new version increments, supersedes, old object unchanged | `version = prev+1`, `supersedes_id = prev.id`, old object bytes intact | `service_test.go:281-315` `TestStoreVersionsANewUploadOverThePreviousDocument` — asserts `v2.Version==2`, `*v2.SupersedesID==v1.ID`, downloads v1's URL after v2 exists and compares to original v1 bytes; DB-level uniqueness of the chain also in `documents_migration_test.go:73-100` `TestDocumentsSupersedesIDIsUniquePreventingFork` | ✅ PASS |
| AC6: no overwrite/delete operation | `tj_app` role has no UPDATE/DELETE path | `service_test.go:318-338` `TestServiceProvidesNoOverwriteOrDeleteOperation` (via the app pool) + `documents_migration_test.go:104-130` `TestDocumentsGrantsTheApplicationRoleOnlyInsertAndSelect` (asserts `permission denied` from Postgres directly) | ✅ PASS |
| AC7: metadata-write failure after upload removes only the new object | object count unchanged (not "an error was returned") | `service_test.go:343-360` `TestStoreRemovesOnlyTheJustCreatedObjectWhenMetadataWriteFails` — forces failure via a nonexistent `Supersedes` id, asserts `errors.Is(err, ErrNotFound)` **and** `objectCount(after) == objectCount(before)` | ✅ PASS |
| AC8: `document.create` audited in the same transaction | audit row with matching action/entity/actor | `service_test.go:363-385` `TestStoreAuditsDocumentCreate` — reads `audit_log` directly, asserts `action`, `entity_type`, `entity_id`, `actor_user_id` | ✅ PASS |
| AC9: missing permission ⇒ forbidden, no storage write | `authz.ErrForbidden`, zero objects, zero rows | `service_test.go:388-408` `TestStoreRequiresTheCallerSuppliedPermissionAndWritesNothingWithoutIt` — asserts error, `objectCount` unchanged, `SELECT count(*) FROM documents WHERE owner_id='l-9'` is 0 | ✅ PASS |
| AC10: extension not allowed rejected before reading content | error `document_extension_not_allowed`, content never read | `service_test.go:411-434` `TestStoreRejectsAnExtensionNotInTheAllowListBeforeReadingContent` — uses a `panicReader` (`service_test.go:173-180`) that fails the test if `Read` is ever called; also proves `.PDF` (uppercase) is accepted | ✅ PASS |
| AC11: extension/detected-type mismatch rejected | error `document_type_mismatch` | `service_test.go:437-449` `TestStoreRejectsWhenTheExtensionDoesNotMatchTheDetectedContentType` (`.pdf` filename, PNG bytes) | ✅ PASS |
| AC12: status born `ACTIVE`, no transition operation | `status == "ACTIVE"` in the returned value and the row | `service_test.go:452-471` `TestStoreSetsStatusToActive` + `documents_migration_test.go:49-69` `TestDocumentsStatusDefaultsToActive` (DB default, no status column passed) | ✅ PASS |

### DOC-02: Acesso a documentos por URL assinada

| Criterion | Spec-defined outcome | `file:line` + assertion | Result |
| --- | --- | --- | --- |
| AC1: presigned URL (300s) + `document.access` audited | working URL, audit row present | `service_test.go:474-504` `TestAccessURLReturnsAPresignedURLAndAuditsAccess` (env configured with `URLTTL: 5*time.Minute` = 300s, `service_test.go:75`) — downloads via the URL and compares bytes, reads `audit_log` for `document.access` | ✅ PASS |
| AC2: missing permission ⇒ forbidden, no URL generated | `authz.ErrForbidden`, `PresignGet` never called | `service_test.go:507-527` `TestAccessURLRequiresTheCallerSuppliedPermissionAndGeneratesNoURL` — uses `spyStorage.presignCalls` counter (`service_test.go:45-53`) to prove zero calls | ✅ PASS |
| AC3: unknown id ⇒ not-found, no URL | `documents.ErrNotFound`, `PresignGet` never called | `service_test.go:530-542` `TestAccessURLReturnsNotFoundForANonexistentIDAndGeneratesNoURL` | ✅ PASS |
| AC4: TTL passed unchanged; `X-Amz-Expires` matches exactly | decoded query param equals the requested TTL | `api/internal/platform/storage/s3_test.go:71-93` `TestS3PresignGetEncodesTheRequestedTTLInTheURL` — `ttl=42s`, asserts `parsed.Query().Get("X-Amz-Expires") == "42"` | ✅ PASS |
| AC5: object without valid signature rejected | non-200 status from the real storage | `storage/s3_test.go:98-114` `TestS3RejectsAnUnsignedRequest` (through the adapter) + `testutil/s3_test.go:70-88` `TestSharedS3RejectsAnUnsignedRequest` (against the raw emulator, proving the chosen emulator itself enforces it) | ✅ PASS |
| AC6: never resolves "current version"; always exact id | requesting v1's id after v2 exists still returns v1's bytes | `service_test.go:546-574` `TestAccessURLNeverResolvesTheCurrentVersionOnItsOwn` | ✅ PASS |

### DOC-03: Configuração do storage

| Criterion | Spec-defined outcome | `file:line` + assertion | Result |
| --- | --- | --- | --- |
| AC1: any missing S3 var (when enabled) ⇒ non-zero exit naming it, no secret printed | error names the exact variable; secret never appears in the error text | `api/internal/config/config_env_test.go:140-158` (table cases naming `S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`) + `:183-195` `TestLoadErrorsNeverContainSecrets` + `:199-237` `TestLoadStorageEnabledReadsAllVariablesAndNeverLeaksTheSecretKey` (asserts the real secret string never appears in `err.Error()`) | ✅ PASS |
| AC2: `DOCUMENT_URL_TTL_SECONDS <= 0` ⇒ refuse to start | `Load()` returns an error naming the variable | `config_env_test.go:159-160` (table cases "validade da URL zero" / "validade da URL negativa") | ✅ PASS |
| AC3: `S3_USE_PATH_STYLE=true` ⇒ client addresses by path | client built with path-style addressing | `api/internal/platform/storage/s3.go:48-51` (`o.UsePathStyle = cfg.UsePathStyle`), exercised with `UsePathStyle: true` in every `storage`/`documents` integration test against Garage (path-style-only emulator — see `testutil/s3.go:210`); config-level default asserted in `config_env_test.go:65-67` | ⚠️ Indirect — no dedicated test toggles the flag `false` against a virtual-hosted endpoint to show the two paths diverge; the boolean's effect is proven only by every test succeeding with it `true`. Trivial pass-through (one line, no branching logic) — low risk, but evidence-or-zero says flag it rather than silently call it a full PASS |

### DOC-04: Consulta de documentos por dono

| Criterion | Spec-defined outcome | `file:line` + assertion | Result |
| --- | --- | --- | --- |
| AC1: every version for owner, newest first, no URL generated | `[v2, v1]` order; `presignCalls` unchanged | `api/internal/platform/documents/list_test.go:17-61` `TestListByOwnerReturnsEveryVersionNewestFirstWithoutGeneratingAURL` — stores v1, v2 (supersedes v1) and a document under a different owner; asserts `docs[0].ID==v2.ID`, `docs[1].ID==v1.ID`, `len(docs)==2` (the other owner's doc excluded), `presignCalls` unchanged | ✅ PASS |
| AC2: missing permission ⇒ forbidden, zero rows | `authz.ErrForbidden`, `len(docs)==0` | `list_test.go:64-85` `TestListByOwnerRequiresTheCallerSuppliedPermissionAndReturnsNoRows` | ✅ PASS |
| AC3: invalid `owner_type` (empty / >100 chars / bad format) ⇒ `document_owner_type_invalid`, no query, no fixed catalog | error before DB access; exact length boundary; arbitrary valid module names accepted | `list_test.go:91-112` `TestListByOwnerRejectsAnInvalidOwnerTypeWithoutQueryingTheDatabase` (proven via an already-cancelled context — a real query would surface a context-cancelled error, not `ErrOwnerTypeInvalid`) + `list_test.go:167-193` `TestListByOwnerEnforcesTheOwnerTypeLengthLimitExactly` (100 accepted, 101 rejected) + `list_test.go:197-208` `TestListByOwnerAcceptsAnyTechnicallyValidOwnerTypeWithNoFixedCatalog` | ✅ PASS |
| AC4: no audit entry for listing | `audit_log` count unchanged | `list_test.go:115-145` `TestListByOwnerDoesNotRecordAnAuditEntry` | ✅ PASS |

**Status**: ✅ All 25 numbered ACs covered with spec-matching assertions; 1 indirect-evidence note (DOC-03 AC3) that does not change the pass/fail outcome given the code's triviality, but is recorded rather than silently rounded up to a full PASS.

---

## AD-015 / Special-context checks (explicit)

1. **DOC-02.4 scope (no emulator-expiry claim)**: confirmed. `storage/s3_test.go` asserts only the encoded `X-Amz-Expires` value (`s3_test.go:71-93`); `testutil/s3_test.go:58-67` carries a comment explicitly recording that no test here claims the emulator rejects an expired URL, consistent with spec.md's Assumption and AD-015. No such test exists anywhere in the diff (grepped for "expir" across the documents/storage/testutil packages — the only hits are TTL pass-through, never emulator-enforced expiry).
2. **Parameterized permission**: confirmed. `grep -rn 'authz\.Permission("'` and a scan of `service.go`/`list.go` found zero hardcoded permission-string constants in the package; `documents.Authorizer` (`service.go:102-104`) is a one-method narrow interface (`Require(ctx, Principal, Permission) error`), not a dependency on `authz`'s concrete catalog type. `StoreInput`, `AccessInput`, `ListInput` all carry `Actor`/`RequiredPermission` as caller-supplied fields.
3. **No fixed `owner_type` catalog**: confirmed. `list.go:22` — `ownerTypeFormat = regexp.MustCompile(`^[a-z0-9_]+\.[a-z0-9_]+$`)`, purely technical; `list_test.go:197-208` proves an unregistered module name (`modulo_futuro.recurso_novo`) is accepted.
4. **"Current version" not the service's job**: confirmed. `AccessURL` (`service.go:268-295`) takes `in.ID` and queries `WHERE id = ?` only — no "latest" resolution path exists in the file. `TestAccessURLNeverResolvesTheCurrentVersionOnItsOwn` (`service_test.go:546-574`) proves it behaviorally.
5. **`tj_app` grants**: confirmed at both levels. Migration (`000005_documents.up.sql:38`): `GRANT INSERT, SELECT ON documents TO tj_app;` — no UPDATE/DELETE. Schema-level proof: `documents_migration_test.go:104-130` (`permission denied` from Postgres). Service-level proof: `service_test.go:318-338` (same, through the app connection pool the service actually uses).
6. **Orphan cleanup proof**: confirmed as an object-count assertion, not just "an error was returned". `service_test.go:357-359`: `if after := objectCount(t, e); after != before { t.Errorf(...) }` — this is a real count of `ListObjectsV2` against the bucket (`service_test.go:141-154`), before and after the forced failure.
7. **CI ran the Garage suite for real**: confirmed via `gh pr checks 38` — job `api` passed in 2m35s. The workflow (`.github/workflows/ci.yml:113-126`) runs `go vet -tags=integration ./...`, `golangci-lint --build-tags=integration`, `go test ./...`, `go test -tags=integration ./...`, `go build ./...`, `govulncheck` on `ubuntu-latest` — the same command from `tasks.md`'s Gate Check table, executed on a real CI runner, not just claimed by the commit message.

---

## Discrimination Sensor

Scratch worktree: `git worktree add ../tj-scratch-verify HEAD` (removed with `git worktree remove --force` after each round; real tree confirmed clean via `git status --porcelain` before, and after, all three rounds — no `git stash` used).

| # | File:line | Mutation | Test run | Killed? |
| - | --------- | -------- | -------- | ------- |
| 1 | `api/internal/platform/documents/service.go:189` | Replaced `if err := s.Authz.Require(...); err != nil { return ..., err }` with `_ = s.Authz.Require(...)` (permission check result discarded — `Store` proceeds regardless of authorization) | `go test -tags=integration ./internal/platform/documents/... -run TestStoreRequiresTheCallerSuppliedPermissionAndWritesNothingWithoutIt` | ✅ Killed — `err = <nil>, esperado authz.ErrForbidden`; object and row were written |
| 2 | `api/internal/platform/storage/s3.go:76` | `s3.WithPresignExpires(ttl)` → `s3.WithPresignExpires(ttl+time.Hour)` (TTL silently altered before reaching the SDK) | `go test -tags=integration ./internal/platform/storage/... -run TestS3PresignGetEncodesTheRequestedTTLInTheURL` | ✅ Killed — `X-Amz-Expires = "3642", esperado "42"` |
| 3 | `api/internal/platform/documents/list.go:97` | `ORDER BY uploaded_at DESC` → `ORDER BY uploaded_at ASC` | `go test -tags=integration ./internal/platform/documents/... -run TestListByOwnerReturnsEveryVersionNewestFirstWithoutGeneratingAURL` | ✅ Killed — got `[v1, v2]`, expected `[v2, v1]` |

**Sensor depth**: lightweight (3 targeted mutations, spanning permission enforcement, TTL pass-through, and ordering — three of the feature's highest-risk behavioral contracts)
**Result**: 3/3 killed — PASS ✅
**Isolation check**: `git status --porcelain` on the real working tree was empty before the sensor run and empty after `git worktree remove --force ../tj-scratch-verify` — confirmed via direct command, not assumed.

---

## Code Quality

| Principle | Status |
| --- | --- |
| Minimum code | ✅ — no speculative abstractions; `Storage` interface has exactly the 3 methods `documents` needs |
| Surgical changes | ✅ — outside the new packages, only `config.go` (additive block), `identity_migration_test.go` (2 lines, a down-migration ordering fix forced by the new FK on `users`), `.env.example`, `go.mod`/`go.sum` (new S3 SDK deps) and the spec/design/tasks/STATE docs were touched |
| No scope creep | ✅ — no HTTP endpoints, no provider decision, no retention/purge, no status-transition workflow — all correctly left out per spec.md's Out of Scope table |
| Matches patterns | ✅ — `Authorizer`/`Auditor` narrow interfaces mirror the existing `identity/app` and `platform/audit` convention; `conn(ctx, db)` mirrors `WithTx`/`TxFrom` usage elsewhere |
| Spec-anchored outcome check (asserted values match spec) | ✅ — see AC table; every numbered AC's test asserts the literal spec-defined value (exact error sentinel, exact TTL, exact order, exact grant set) |
| Per-layer Coverage Expectation met (domain 1:1 ACs; routes happy+edge+error) | ✅ — no routes in scope (feature is platform-only, HTTP explicitly out of scope); service/repository layers have 1:1 AC coverage plus permission-denied and transaction paths |
| Every test maps to a spec requirement — no unclaimed tests | ✅ — every test function's comment cites a DOC-0X.N id or a plainly-scoped edge case (empty-owner listing, exact length boundary) |
| Documented guidelines followed | `tasks.md`'s Test Coverage Matrix (unit vs integration split, `-tags=integration` for real Postgres/S3) and `CLAUDE.md` (money in cents — N/A here, no money column; migrations only via `golang-migrate`, no `AutoMigrate`; RBAC checked in the use case) — all followed |

No unrelated code was "improved". No dead code was removed beyond what the change orphaned. `boolEnv` and `positiveInt` in `config.go` reuse the existing helper pattern rather than duplicating validation logic.

---

## Edge Cases

- [x] Presigned URL validity ≤ 0 ⇒ configuration error at startup — `config_env_test.go:159-160`
- [x] Object requested without a valid signature is rejected — `storage/s3_test.go:98-114`, `testutil/s3_test.go:70-88`
- [ ] Same file stored twice as two unrelated documents ⇒ two distinct keys — **not directly tested**. Guaranteed structurally (`doc.ID: uuid.New()` per call, `service.go:213`, no content-based dedup path exists anywhere in `Store`), and `list_test.go:39-45` incidentally stores byte-identical content (`pdfBytes()`) as a second, unrelated document without asserting key distinctness. No test asserts this outcome directly.
- [ ] Storage unreachable during a store operation ⇒ storage error returned, no metadata written — **partially tested**. `storage/s3_test.go:163-181` `TestS3PutWrapsAnUnreachableEndpointInErrWrite` proves the *adapter* fails with `ErrWrite` against an unreachable endpoint, but no `documents` package test injects a failing `Storage` double into `Service.Store` and asserts zero rows were written at the *service* level for this scenario (the metadata-write-failure test, AC7, exercises a different failure point — the transaction, after a successful upload). Code order (`service.go:220-224`: `Storage.Put` strictly precedes `database.WithTx`) makes the guarantee hold today, but no test would catch a regression that reordered these two calls.
- [ ] Two callers with different `RequiredPermission` values for the same kind of operation ⇒ exactly the one passed for that call is checked — **not directly tested with two differing calls**, but structurally guaranteed: `Service` (`service.go:136-143`) has no field that stores a permission across calls, and `RequiredPermission` is read fresh from `StoreInput`/`AccessInput`/`ListInput` on every call (`service.go:189`, `272`; `list.go:87`). No regression test specifically proves non-interference across two back-to-back calls with different permissions.

---

## Gate Check

- **Gate command**: `cd api && go vet ./... && go test -tags=integration ./...` (also ran `go vet -tags=integration ./...` and `go build ./...` per the task's explicit ask)
- **Result**: `go vet ./...` — clean, no output. `go vet -tags=integration ./...` — clean, no output. `go build ./...` — clean. `go test ./...` (no tag) — all packages `ok`. `go test -tags=integration ./...` — **701 subtests succeeded, 0 failed** in a clean run (counted via `go test -tags=integration -v ./...` grepped for the passing- and failing-subtest markers: 701 / 0).
- **Flake note (environment, not code)**: two earlier full-suite runs of `go test -tags=integration ./...` (default parallelism) intermittently failed several *pre-existing, unrelated* tests in `platform/database` and/or `platform/testutil` with `rootless Docker is not supported on Windows, failed to create Docker provider` — a testcontainers-go provider-detection race under concurrent package startup on Windows Docker Desktop, not specific to this feature. Evidence this is environmental: (a) the failing tests differed between the two runs (once `testutil`, once `database`), (b) `documents` and `storage` (the packages this feature actually adds) passed cleanly in every run including the flaky ones, (c) running `go test -tags=integration ./internal/platform/testutil/...` alone passed 100%, (d) running the whole suite with `go test -p 1 -tags=integration ./...` (serialized packages) passed 100% with zero failures, and (e) the same suite passed in real CI on `ubuntu-latest` for PR #38 (`gh pr checks 38` → `api` job pass, 2m35s) — Linux CI runners do not hit this Windows-specific testcontainers path.
- **Test count before feature** (in scope packages): `documents` package: 0 (did not exist). `storage` package: 0 (did not exist). `testutil/s3.go` + `s3_test.go`: 0 (did not exist). `database` package: migration tests existed for `audit_log`, `identity`/`users`/`sessions`/`permissions`/`admin_memberships`, `password_reset` (pre-existing, untouched in content).
- **Test count after feature**: `documents`: 23 new test functions (16 in `service_test.go` + 7 in `list_test.go`). `storage`: 5 new (`s3_test.go`). `testutil`: 2 new (`s3_test.go`) plus the new `s3.go` helper. `database`: 6 new (`documents_migration_test.go`) + 2 existing tests in `identity_migration_test.go` modified only to add a down-migration step for the new FK (no assertion weakened). `config`: 2 new top-level test functions plus ~10 new table-driven cases across existing tests.
- **Delta**: +38 new test functions (well above every task's stated minimum: T1≥3, T3≥2, T4≥4, T5≥4, T6≥15 [16 delivered], T7≥6 [7 delivered]). No test was deleted, skipped, or weakened anywhere in the diff.
- **Skipped tests**: none.
- **Failures**: none attributable to this feature's code (see flake note above).

---

## Fix Plans

None required to reach PASS. The Edge Cases items left unchecked above are minor, structurally-guaranteed gaps in *test* coverage (not in behavior) — recorded as a lesson below rather than blocking the feature, consistent with their Minor severity (infrastructure code, guaranteed by simple, already-reviewed control flow, not by anything fragile).

---

## Requirement Traceability Update

| Requirement | Previous Status | New Status |
| --- | --- | --- |
| DOC-01 | Implemented | ✅ Verified |
| DOC-02 | Implemented | ✅ Verified |
| DOC-03 | Implemented | ✅ Verified (1 indirect-evidence note on AC3, non-blocking) |
| DOC-04 | Implemented | ✅ Verified |

---

## Summary

**Overall**: ✅ Ready

**Spec-anchored check**: 25/25 numbered ACs matched their spec-defined outcome with direct `file:line` evidence; 1 indirect-evidence note (DOC-03 AC3, trivial boolean pass-through)

**Sensor**: 3/3 mutations killed (permission bypass in `Store`, TTL tampering in `PresignGet`, `ORDER BY` direction in `ListByOwner`) — real worktree confirmed unchanged before and after

**Gate**: `go vet`, `go vet -tags=integration`, `go build`, `go test`, and `go test -tags=integration` all pass; 701/701 subtests green in a clean serialized/CI run; two intermittent failures observed under default local parallelism traced conclusively to a Windows-only testcontainers/Docker-Desktop race unrelated to this feature's code (isolated re-runs and real CI both green)

**What works**: Immutable, versioned S3-protocol storage with correct chained `supersedes_id`; 10 MiB / extension / content-type / extension-mismatch validation in the right order and before any metadata write; orphan-object cleanup proven by an actual object-count check, not just an error; `document.create`/`document.access` audit entries in the same transaction as the write; zero hardcoded permissions in `platform/documents` (AD-015 upheld both structurally and by test); `tj_app` has no UPDATE/DELETE, proven at both the Postgres and service layers; `ListByOwner` has no fixed `owner_type` catalog and never audits or signs a URL; DOC-02.4's reformulated scope (TTL pass-through only, no claim about emulator-enforced expiry) is respected to the letter; CI genuinely ran the Garage-based integration suite on PR #38.

**Issues found**: Three Edge Cases from spec.md (same-file-twice distinct keys; storage-unreachable-during-Store leaves no metadata at the *service* level; permission non-interference across two differently-permissioned calls) are behaviorally guaranteed by the current code but have no dedicated test that would catch a future regression in that specific guarantee. Not blocking (Minor severity, infra code, simple control flow) — recorded as a lesson for future platform-layer features.

**Next steps**: None required to close this feature. Optional future hardening (not a gate item): add a `documents`-level test with a failing `Storage` double for the "unreachable during Store" edge case, and a same-content/no-supersedes test asserting distinct storage keys, the next time this package is touched.
