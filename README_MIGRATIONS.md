Flyway migration notes
---------------------

This project uses Flyway CLI as the canonical migration tool. Use the provided Makefile targets:

- `make flyway-migrate` — apply migrations (will baseline on first run if schema non-empty).
- `make flyway-info` — show Flyway info.
- `make flyway-history` — show `flyway_schema_history` table contents.
- `make db-drop` — destructive; runs `flyway clean`.

Place Flyway-style migrations in `internal/infrastructure/database/migrations` with names like `V1__desc.sql`.

If you previously used `golang-migrate`, its artifacts (`schema_migrations`) were removed and Flyway's
`flyway_schema_history` is the canonical history table now.
