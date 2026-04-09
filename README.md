# Go REST API Pro

Resumo
-----------------
Projeto exemplo de uma API REST em Go, com foco em boas práticas de conteinerização, migrações versionadas (Flyway), logs estruturados e um fluxo de desenvolvimento que facilita executar, migrar e depurar localmente com `docker compose` e `Makefile`.

Motivação
-----------------
- Atualizar toolchain para Go 1.25.9 para compatibilidade com dependências.
- Substituir migrações em execução pelo binário por uma solução robusta com Flyway (execução via container) e manter o histórico em `flyway_schema_history`.
- Tornar o contêiner de runtime mais seguro e leve (usuário não-root, etapa multi-stage).
- Adicionar healthchecks, retry de conexão com o DB e logs estruturados (zerolog) para observabilidade.

O que foi implementado
-----------------
- Multi-stage `Dockerfile` com builder em `golang:1.25.9` e runtime baseado em `alpine`.
- `Makefile` com targets úteis: `deps`, `up`, `down`, `up-api`, `rebuild`, `flyway-migrate`, `flyway-info`, `flyway-history`, `flyway-clean`.
- Migrações Flyway em `internal/infrastructure/database/migrations/` (V1..V4 aplicadas).
- Serviço `migrate` no `docker-compose.yml` que executa as migrações antes da API iniciar.
- Retry na conexão com PostgreSQL em `internal/infrastructure/database/connection.go` (variáveis: `DB_MAX_RETRIES`, `DB_RETRY_DELAY`).
- Stub de compatibilidade `RunMigrations` mantido para evitar que a aplicação quebre por referência removida.
- Logs estruturados com `zerolog` controlados por `LOG_LEVEL`.
- Healthchecks: `pg_isready` para Postgres e endpoint HTTP `/health` para a API.

Pré-requisitos
-----------------
- Docker (>= 20.x) e a extensão/plug-in `docker compose` (comando `docker compose`).
- Make (opcional, facilita comandos encadeados).
- (Opcional) `psql` para inspeção manual do banco.

Arquivos importantes
-----------------
- Código da API: [cmd/api/main.go](cmd/api/main.go)
- Conexão com DB: [internal/infrastructure/database/connection.go](internal/infrastructure/database/connection.go)
- Migrações Flyway: [internal/infrastructure/database/migrations/](internal/infrastructure/database/migrations/)
- Makefile: [Makefile](Makefile)
- Docker Compose: [docker-compose.yml](docker-compose.yml)
- Dockerfile: [Dockerfile](Dockerfile)
- Variáveis de ambiente de exemplo: [.env.example](.env.example)

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

4. Subir todo o stack (Postgres + Flyway migrate + API):

```bash
make up
```

Observações:
- O target `make up` levanta os serviços com `-d --build` e aguarda o serviço `migrate` (Flyway) aplicar as migrações antes de iniciar a API.
- Use `make up-api` se quiser subir somente a API (útil para desenvolvimento rápido quando o DB já está rodando).

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

---
Gerado em: 9 de abril de 2026
