# go-rest-api-cqrs-event-sourcing

## 🧠 Overview

Template profissional em Go utilizando:

- CQRS (Command Query Responsibility Segregation)
- Event Sourcing
- Arquitetura Hexagonal
- Kafka (Event Bus)
- Redis (Read Model - Hash + Index)
- Postgres (Event Store)
- Gin (HTTP API)

---

## 🏗️ Arquitetura
cmd/api/

internal/
domain/
user/

application/
command/
query/

infrastructure/
database/postgres/
cache/redis/
messaging/kafka/

interfaces/http/gin/

projection/


---

## 🔁 Fluxos

### WRITE (Command)
HTTP → Gin → Command Handler → Aggregate → Events → Postgres → Kafka


---

### READ (Query)
Kafka → Consumer → Projection → Redis
HTTP → Query Handler → Redis


---

## 🧱 Componentes

### 🐘 Postgres (Event Store)

Tabela principal:

```sql
CREATE TABLE events (
  id SERIAL PRIMARY KEY,
  aggregate_id TEXT NOT NULL,
  type TEXT NOT NULL,
  payload JSONB NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

✔ Fonte da verdade
✔ Armazena eventos, não estado

📨 Kafka (Event Bus)
Producer publica eventos após persistência
Consumer consome eventos para projeções

✔ Desacoplamento
✔ Escalabilidade

⚡ Redis (Read Model)

Estrutura:
user:{id} → HASH

✔ Leitura ultra rápida
✔ Dados denormalizados

🧩 Estrutura de Código
Domain
Aggregates
Events
Regras de negócio

Application
Commands (write)
Queries (read)

Infrastructure
Postgres (event store)
Kafka (producer + consumer)
Redis (read model)

Interfaces
HTTP API com Gin

🐳 Docker Compose
Serviços:
postgres
redis
zookeeper
kafka

▶️ Execução (local)

1. Prepare as variáveis de ambiente:

```bash
cp .env.example .env
# Edit .env as needed (DB credentials, kafka, redis)
```

2. Subir os serviços via Docker Compose:

```bash
make up
```

3. Aplicar migrações (usando Flyway container):

```bash
make migrate      # alias para make flyway-migrate
```

4a. Rodar a API em container (recomendado):

```bash
make up-api
```

4b. Ou rodar localmente (Go runtime):

```bash
make deps
make run
```

Outros alvos úteis:

```bash
make down         # derruba containers e volumes
make ps           # ver containers
make logs         # logs do compose
```

⚠️ Consistência
Postgres → consistente
Redis → eventual
👉 Isso é esperado em CQRS

Regras importantes
❌ Nunca atualizar Redis diretamente via API
✔ Sempre via eventos (Kafka)
✔ Postgres é a única fonte da verdade

🧠 Conclusão
Esse projeto representa uma base sólida para:

Sistemas distribuídos
Alta escalabilidade
Alta performance de leitura
Baixo acoplamento

👉 Arquitetura nível fintech / enterprise

🚀 Evoluções Futuras (Não desenvolver nesse momento)
Versionamento de eventos
Idempotência
Snapshotting
Observabilidade (OpenTelemetry)
Retry / Dead Letter Queue
Multi-aggregate (User, Order, Payment)