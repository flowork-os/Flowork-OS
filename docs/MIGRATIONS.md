# Migration contract — CODE vs DATA

> **The law:** an update replaces CODE; it must never drop, rewrite, or lose user DATA.
> New schema and new seed content merge in **additively** and **idempotently**. This is the
> guarantee behind auto-update: *"DB lama selamat, DB baru masuk tanpa merusak."*

Both engines already follow this. This doc is the contract every future change must keep.

## Rules (forward-only, additive, idempotent)

1. **New table** → `CREATE TABLE IF NOT EXISTS …`. Never `DROP TABLE`.
2. **New column** → `ALTER TABLE … ADD COLUMN … NOT NULL DEFAULT …`, guarded so it runs once:
   - Router: a registered `Migration{ID, Name, SQL}` (see below).
   - Agent: `if !columnExists(table, col) { ALTER … }` (e.g. `EnsureTaskSchema` in
     `internal/floworkdb/tasks.go`).
   Never `DROP COLUMN`, never rename a column in place (that loses data on SQLite).
3. **Seed / default content** → `INSERT OR IGNORE`, `ON CONFLICT … DO UPDATE` only for
   machine-owned fields, or seed-only-when-empty. A user edit via the GUI must never be
   overwritten by a re-seed. (See `agent/main.go` prompt seed, `promptseed.go`.)
4. **Deletes** are user-initiated CRUD only (`DELETE … WHERE id=?`). No upgrade step deletes rows.
5. **DATA lives apart from CODE.** User DBs (`*.db`, `router/brain/*.sqlite`, agent workspaces)
   are gitignored and live on the DATA partition / app data dir. Updates swap the ROOT/binaries,
   never the DATA store. (Mirrors the OS A/B mechanism: ROOT slots flip, DATA partition untouched.)

## Router — formal registry

`router/internal/store/migrate.go` (LOCKED — owner approval to modify): ordered
`Migration{ID, Name, SQL}` registered from `init()` in sibling files
(`mesh_stack_migrations.go`, `llm_pricing_policy_migrations.go`); `applyMigrations()` runs every
migration with `ID >` the highest recorded in the `schemaMigrations` table, inside one
transaction. Adding a change = append a new `Migration` with the next `ID`. Never edit an
already-shipped migration's SQL.

## Agent — guarded ensure

`internal/floworkdb` and `internal/agentdb` create tables with `CREATE TABLE IF NOT EXISTS` and
add later columns with a `columnExists`/`PRAGMA table_info` guard before `ALTER … ADD COLUMN`.
Adding a change = a new guarded `ALTER` in the relevant `Ensure*Schema`.

## Proof

`agent/internal/floworkdb/migration_zeroloss_test.go` (`TestMigrationZeroLoss`) builds a DB at an
**older** schema with real user data, runs the current upgrade path, and asserts: every old row
survives intact, the new columns are added with their defaults, and a second upgrade is a no-op.

```sh
go test ./agent/internal/floworkdb/ -run TestMigrationZeroLoss -v
```

When you add a migration, extend this test (or add a sibling) with the old→new case for it.
