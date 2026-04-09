Database helpers
----------------

This package contains DB connection helpers. Migrations are handled externally
via Flyway (`make flyway-migrate`). The old `golang-migrate` runtime integration
was removed; a no-op `RunMigrations` stub remains for compatibility.
