## Why

O repositório atual usa `database/sql` puro com queries SQL escritas à mão e placeholders posicionais do PostgreSQL (`$1, $2...`). GORM elimina esse boilerplate, oferece migrations automáticas, e é o padrão adotado no ecossistema Go para projetos que precisam de evolução de schema sem migrations manuais.

## What Changes

- Substituição de `PostgresRepository` (baseado em `database/sql`) por implementação equivalente usando GORM
- Substituição de `AutoMigrate` no lugar do schema SQL inline (`CREATE TABLE IF NOT EXISTS`)
- Remoção das funções utilitárias manuais (`nullableStr`, `nullableFloat`, `nullableTime`) — GORM lida com ponteiros nativamente
- Adição de model GORM (`Payment`) mapeado ao domínio existente
- Atualização de `go.mod` para incluir `gorm.io/gorm` e `gorm.io/driver/postgres`
- Remoção de `github.com/jackc/pgx/v5` como driver direto (substituído pelo driver GORM para Postgres)

## Capabilities

### New Capabilities

- `gorm-payment-repository`: Implementação do repositório de pagamentos usando GORM, com model, AutoMigrate e queries via ORM

### Modified Capabilities

*(nenhuma spec existente tem requisitos que mudam — a troca é puramente de implementação)*

## Impact

- **Arquivo removido**: `internal/adapters/out/database/postgres_repository.go`
- **Arquivo criado**: `internal/adapters/out/database/payment_model.go` (model GORM)
- **Arquivo criado**: `internal/adapters/out/database/gorm_repository.go` (implementação do repositório)
- **Dependências**: adiciona `gorm.io/gorm`, `gorm.io/driver/postgres`; remove `github.com/jackc/pgx/v5`
- **Interface de domínio**: sem alteração — `PaymentRepository` continua igual
- **`main.go`**: troca `NewPostgresRepository` por `NewGORMRepository`
