# Database migrations

Schema migrations live in `internal/database/migration/migrations`. They are embedded in the migration binary, so the SQL version that is executed always travels with the application source.

## Run pending migrations

From the repository root, configure the same database variables used by the API (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, and `DB_SSLMODE`), then run:

```bash
go run ./cmd/migrate
```

The command creates `schema_migrations` when needed, then applies each pending `*.up.sql` file in lexical order. A migration and its version record are committed in one PostgreSQL transaction. PostgreSQL advisory locking prevents two concurrent migration commands from applying the same version.

The API does **not** run migrations on startup. Schema changes stay an explicit deployment step rather than an application side effect.

## Initial Category baseline

`000001_create_category.up.sql` represents the current `category` table used by the repository. It uses `CREATE TABLE IF NOT EXISTS` because Category already exists in the current database.

On that existing database, the first command records `000001_create_category` as applied without replacing the table. Before that first run, confirm that the existing table has the expected columns and types. The migration runner records versions; it intentionally does not attempt to infer or repair a pre-existing schema.

## Add a new migration

1. Add one immutable, forward-only file in `internal/database/migration/migrations` using the next zero-padded number, for example `000002_add_category_slug.up.sql`.
2. Put the PostgreSQL SQL needed to evolve the schema in that file.
3. Run `go run ./cmd/migrate` in each environment as an explicit deployment step.
4. Never edit a migration that has reached a shared environment. Create a later corrective migration instead.

Migrations are intentionally forward-only: there are no automatic down migrations, since rolling back a schema can destroy or reinterpret existing data. Correct an applied change with a new forward migration or restore a database backup through the operational recovery process.
