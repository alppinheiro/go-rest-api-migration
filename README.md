# Go REST API CQRS Event Sourcing

Resumo
-----------------
Projeto exemplo de uma API REST em Go com CQRS, Event Sourcing e arquitetura hexagonal mínima para o agregado `User`. O fluxo de escrita persiste eventos no Postgres, publica no Kafka e a projeção atualiza o read model no Redis. O fluxo de leitura consulta apenas o Redis.

Motivação
-----------------
- Evoluir a base existente para a arquitetura descrita em `instructions.md` sem introduzir pipeline, testes de integração ou observabilidade adicional neste momento.
- Materializar um primeiro fluxo vertical completo de comando e consulta para servir de base para os próximos aggregates.

O que foi implementado
-----------------
- Multi-stage `Dockerfile` com builder em `golang:1.25.9` e runtime baseado em `alpine`.
- `Makefile` com targets úteis: `deps`, `up`, `down`, `up-api`, `run`, `migrate`, `flyway-info`, `flyway-history`, `flyway-clean`, `ps`, `logs`.
- Estrutura inicial de arquitetura hexagonal em `internal/domain`, `internal/application`, `internal/infrastructure` e `internal/interfaces`.
- Aggregate `User` com evento de domínio `user.created`.
- Event store em Postgres com tabela `events` criada pela migration `V5__create_events_table.sql`.
- Publisher Kafka para eventos de domínio e projector que consome `user-events` e atualiza o read model no Redis.
- Read model `user:{id}` e índice `users:email_index` no Redis.
- Endpoints HTTP implementados:
	- `POST /users`
	- `GET /users/:id`
	- `GET /users?email=...`
	- `GET /health`
- Serviço `migrate` no `docker-compose.yml` que executa as migrações antes da API iniciar.
- Retry na conexão com PostgreSQL em `internal/infrastructure/database/connection.go`.
- Logs estruturados com `zerolog` controlados por `LOG_LEVEL`.

Pré-requisitos
-----------------
- Docker (>= 20.x) e a extensão/plug-in `docker compose` (comando `docker compose`).
- Make (opcional, facilita comandos encadeados).
- (Opcional) `psql` e `redis-cli` para inspeção manual dos stores.

Arquivos importantes
-----------------
- Bootstrap da API: [cmd/api/main.go](cmd/api/main.go)
- Aggregate `User`: [internal/domain/user/user.go](internal/domain/user/user.go)
- Command handler: [internal/application/command/create_user.go](internal/application/command/create_user.go)
- Query handler: [internal/application/query/get_user.go](internal/application/query/get_user.go)
- Projector: [internal/application/projection/user_projector.go](internal/application/projection/user_projector.go)
- Event store: [internal/infrastructure/database/event_store.go](internal/infrastructure/database/event_store.go)
- Read model Redis: [internal/infrastructure/cache/redis/user_read_model.go](internal/infrastructure/cache/redis/user_read_model.go)
- Rotas HTTP: [internal/interfaces/http/gin/router.go](internal/interfaces/http/gin/router.go)
- Migrações Flyway: [internal/infrastructure/database/migrations/](internal/infrastructure/database/migrations/)
- Docker Compose: [docker-compose.yml](docker-compose.yml)
- Makefile: [Makefile](Makefile)

Como usar (Local / Desenvolvimento)
-----------------
1. Copie o arquivo de exemplo de ambiente e ajuste se necessário:

```bash
cp .env.example .env
```

2. Instale dependências do sistema (Docker + Make) e certifique-se de que `docker compose` funciona:

```bash
docker compose version
```

3. Preparar dependências (opcional):

```bash
make deps
```

4. Subir todo o stack (Postgres + Redis + Zookeeper + Kafka + Flyway migrate + API):

```bash
make up
```

Observações:
- O target `make up` levanta os serviços com `-d --build` e aguarda o serviço `migrate` aplicar as migrações antes de iniciar a API.
- As portas externas são configuráveis na `.env` (`POSTGRES_PORT`, `REDIS_PORT`, `KAFKA_PORT`, `ZOOKEEPER_PORT`, `API_PORT`).

Exemplo rápido do fluxo:

```bash
curl -X POST http://localhost:8081/users \
	-H 'Content-Type: application/json' \
	-d '{"name":"Andre Luiz","email":"andre@example.com"}'

curl http://localhost:8081/users/<id-retornado>
curl 'http://localhost:8081/users?email=andre@example.com'
```

Executando migrações manualmente (Flyway)
-----------------
As migrações são gerenciadas por Flyway e podem ser executadas via targets do `Makefile`:

```bash
make flyway-migrate   # aplica migrações
make flyway-info      # mostra estado do Flyway
make flyway-history   # mostra histórico completo
make flyway-clean     # limpa schema (CUIDADO: remove objetos do DB)
```

O `Makefile` usa a imagem oficial `flyway/flyway` e monta o diretório `internal/infrastructure/database/migrations/` para que as SQLs sejam detectadas.

Serviço `migrate` (Flyway) — por que e como é usado
-----------------
Onde está definido:
- Serviço `migrate` em [docker-compose.yml](docker-compose.yml).

O que ele faz:
- Executa a imagem oficial `flyway/flyway` e monta `internal/infrastructure/database/migrations/` em `/flyway/sql`.
- Comando padrão usado no Compose:
	- `-url=jdbc:postgresql://postgres:5432/${POSTGRES_DB} -user=${POSTGRES_USER} -password=${POSTGRES_PASSWORD} -baselineOnMigrate=true migrate`
- Cada migração aplicada é registrada na tabela `flyway_schema_history` do banco.

Por que usamos este serviço:
- Mantém a responsabilidade de migração separada do binário da aplicação, garantindo que o esquema seja aplicado de forma reprodutível antes da API iniciar.
- Permite auditar e reverter (via histórico) as alterações de esquema e padroniza execução em diferentes ambientes.

Orquestração com `docker compose`:
- `migrate` depende do serviço `postgres` com healthcheck (`pg_isready`).
- `api` depende de `migrate` usando `service_completed_successfully`, de forma que o Compose só inicia a API após o Flyway terminar com sucesso.

Como executar manualmente:
- Via Makefile (recomendado):

```bash
make flyway-migrate
```

- Diretamente com `docker compose` (executa um container Flyway que aplica migrações):

```bash
docker compose run --rm migrate
```

Comandos úteis para inspecionar o histórico (exemplo genérico):

1) Identifique o nome do container Postgres:

```bash
docker ps
```

2) Execute uma query para ver `flyway_schema_history`:

```bash
docker exec <postgres_container> \
	psql -U <db_user> -d <db_name> -c "SELECT installed_rank, version, description, script, installed_by, installed_on, success FROM flyway_schema_history ORDER BY installed_rank;"
```

3) Inspecione dados da tabela afetada (ex.: `users`):

```bash
docker exec <postgres_container> \
	psql -U <db_user> -d <db_name> -c "SELECT id, name, email, created_at FROM users ORDER BY created_at LIMIT 10;"
```

Observações importantes:
- O flag `-baselineOnMigrate=true` permite adotar um banco existente como ponto de partida (útil ao migrar um DB que já contém objetos).
- A aplicação não executa mais migrações internamente — o arquivo `internal/infrastructure/database/gomigrate.go` é apenas um stub informativo. Se remover o serviço `migrate`, assegure-se de que as migrações serão aplicadas por outro processo (CI, manual ou embutidas no deploy).
- Em ambientes de produção, prefira executar migrações via pipeline controlado (CI/CD) com backups e janela de manutenção, em vez de depender apenas do `docker compose`.

Desenvolvimento sem Docker (rodar a API localmente)
-----------------
1. Configure as variáveis de ambiente no `.env` ou no seu ambiente local.
2. Use `go run` para iniciar a API diretamente (útil para debug):

```bash
go run ./cmd/api
```

Nota: a aplicação possui retry na conexão com o DB; assegure que o Postgres esteja acessível na URL configurada.

Variáveis de ambiente principais
-----------------
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` — conexão com Postgres.
- `DB_MAX_RETRIES` — número máximo de tentativas de conexão (padrão configurado no código).
- `DB_RETRY_DELAY` — tempo em ms entre tentativas.
- `REDIS_ADDR`, `REDIS_PASSWORD` — conexão com Redis/read model.
- `KAFKA_BROKERS`, `KAFKA_TOPIC`, `KAFKA_RETRY_DELAY` — broker, tópico e retry do event bus.
- `LOG_LEVEL` — nível de logs (`debug`, `info`, `warn`, `error`).
- `GIN_MODE` — modo do Gin (`release` em runtime Dockerfile por padrão).

Saúde e observabilidade
-----------------
- Endpoint health: `GET /health` (retorna 200 quando o app está pronto).
- Logs em JSON via `zerolog`.
- Postgres healthcheck configurado em `docker-compose.yml` usando `pg_isready`.

Notas sobre o `Dockerfile`
-----------------
- O `Dockerfile` usa etapa de build com `golang:1.25.9` para garantir compatibilidade com módulos.
- A imagem final é baseada em `alpine` com um usuário não-root criado (UID 1000) para segurança.
- As migrações também são copiadas para o container runtime para facilitar inspeção, porém as migrações efetivas são aplicadas pelo serviço `migrate` do Compose.

Alterações recentes
-----------------
- Removidas as variáveis de ambiente não utilizadas: `RUN_GOMIGRATE` (antes em `.env`) e `RUN_MIGRATIONS` (antes definida no `api` do `docker-compose.yml`). Essas variáveis não são mais lidas pelo código; as migrações devem ser executadas via Flyway (`make flyway-migrate`) ou pelo serviço `migrate` do Compose.

Boas práticas e recomendações
-----------------
- Nunca guarde segredos em `.env` em repositórios públicos; para produção use secret managers (Vault, AWS Secrets Manager, Kubernetes Secrets).
- Adicionar um job de CI que execute `make flyway-migrate` contra um banco temporário e rode testes de integração.
- Considerar embutir migrações no binário para deploys que não usam Flyway (opcional) ou manter Flyway como padrão para ambientes orquestrados.
- Para produção, avaliar imagens distroless e scanning de vulnerabilidades.

Como adicionar uma nova migração
-----------------
1. Crie um arquivo SQL seguindo o padrão Flyway: `V<version>__<description>.sql` dentro de `internal/infrastructure/database/migrations/`.
2. Garanta que o SQL é idempotente quando fizer sentido e que segue as constraints desejadas.
3. Rode localmente para validar:

```bash
make flyway-migrate
make flyway-history
```

Resolução de problemas comuns
-----------------
- Erro: `docker-compose: comando não encontrado` — use `docker compose` (plugin). O `Makefile` tenta detectar automaticamente.
- Erro ao compilar por versão do Go — atualize sua toolchain para Go 1.25.x.
- API não conecta ao DB — verifique `.env`, `DB_HOST` (no Compose, o host do serviço é `postgres`) e aumente `DB_MAX_RETRIES` se necessário.

Próximos passos recomendados
-----------------
- Mover segredos para um secret manager antes de qualquer deploy real.
- Adicionar middleware de request-id e logs por requisição.
- Integrar testes de integração no CI pipeline que executam as migrações e validam contratos.

Como eu posso ajudar mais
-----------------
Se precisar que eu gere um arquivo de CI, scripts para criar DBs de teste, ou que eu aplique essas mudanças direto no repositório, diga qual tarefa prefere que eu faça a seguir.

## Development & Debugging

- **Docker / Compose**: the project uses a named network `app_net` and persistent volumes for core services (`postgres_data`, `redis_data`, `kafka_data`, `zookeeper_data`). Postgres is mounted at `/var/lib/postgresql` (required for Postgres 18+ images).

- **Makefile shortcuts**:
	- `make clean` — remove containers, networks and volumes (`docker compose down -v --remove-orphans`).
	- `make start` — start core services: `postgres`, `redis`, `zookeeper`, `kafka`, `migrate`.
	- `make up` / `make down` — full compose up/down wrappers.
	- `make ps`, `make logs` — inspect services and follow logs.

- **Debugging (VS Code)**: a minimal launch config is available at [/.vscode/launch.json](.vscode/launch.json). Use the `Launch API (debug)` configuration to run with Delve; `dlvToolPath` is set in the config. If you prefer headless Delve, use the `Attach to Delve :2345` configuration.

- **Projection / Read model**: events are written to Postgres; the read model is populated by the projector consuming `user-events` from Kafka and writing to Redis. Ensure `PROJECTION_ENABLED=true` when running the API locally to enable the projector. If Redis is empty, either enable projections (Kafka + projector) or run a reprojection tool to rebuild the read model from the `events` table.

- **Repository hygiene**: `.env` and `.vscode/` are ignored by `.gitignore` — use `.env.example` as the template for local environment variables.

---
Gerado em: 9 de abril de 2026
