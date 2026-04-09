---
name: Migration / Flyway change
about: Notes about migration tool change to Flyway
---

Planned migration tool change: switch canonical migrations to Flyway and remove golang-migrate runtime usage.

Checklist:
- [x] Convert existing migrations to Flyway format
- [x] Add Flyway Makefile targets
- [x] Remove golang-migrate runtime invocation
- [x] Add README_MIGRATIONS.md
