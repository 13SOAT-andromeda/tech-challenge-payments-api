## Context

O repositório de pagamentos (`PostgresRepository`) usa `database/sql` puro com queries SQL escritas à mão. O schema é criado via string SQL inline no `const schema`, e cada operação exige mapeamento manual de colunas com `Scan`. A interface `PaymentRepository` já está bem definida no domínio e não muda.

O model de domínio (`domain.Payment`) usa tipos como `*time.Time` e enums string — ambos compatíveis com as tags GORM.

## Goals / Non-Goals

**Goals:**
- Substituir `database/sql` por GORM mantendo a mesma interface de repositório
- Usar `AutoMigrate` para gerenciar o schema, eliminando o SQL inline
- Separar o model GORM do domain object (anti-corruption layer)
- Remover `pgx/v5` como dependência direta

**Non-Goals:**
- Alterar a interface `PaymentRepository` ou o domínio
- Introduzir migrations versionadas (ex: golang-migrate) — `AutoMigrate` é suficiente para este estágio
- Mudar qualquer comportamento de negócio

## Decisions

### 1. Model GORM separado do domain.Payment

**Decisão:** Criar `internal/adapters/out/database/payment_model.go` com uma struct `paymentModel` com tags GORM, distinta de `domain.Payment`.

**Motivo:** Evita vazar anotações de persistência (`gorm:"..."`) para o domínio. A conversão entre model e domain fica no repositório.

**Alternativa descartada:** Anotar `domain.Payment` diretamente com tags GORM — viola separação de camadas.

---

### 2. Driver: `gorm.io/driver/postgres`

**Decisão:** Usar o driver oficial GORM para Postgres, que usa `pgx` internamente (via `jackc/pgx`).

**Motivo:** Integração nativa com GORM, sem necessidade de gerenciar `*sql.DB` manualmente.

**Alternativa descartada:** Manter `pgx/v5` direto e passar o `*sql.DB` para o GORM — adiciona complexidade sem benefício.

---

### 3. AutoMigrate no lugar de schema SQL manual

**Decisão:** Usar `db.AutoMigrate(&paymentModel{})` na inicialização.

**Motivo:** Garante que o schema evolua automaticamente junto com o model GORM, sem manter SQL sincronizado manualmente.

**Limitação conhecida:** `AutoMigrate` não faz rollback de colunas removidas — aceito neste estágio.

---

### 4. Mapeamento de tipos nullable

**Decisão:** Usar ponteiros (`*string`, `*float64`, `*time.Time`) nas colunas anuláveis do model GORM.

**Motivo:** GORM nativamente converte `nil` para `NULL` com ponteiros, eliminando as funções `nullableStr`, `nullableFloat`, `nullableTime`.

## Risks / Trade-offs

- **AutoMigrate em produção** → Em tabelas com dados, `AutoMigrate` adiciona colunas mas não remove nem renomeia. Mudanças destrutivas precisarão de migration manual. Mitigação: documentar no README que colunas removidas do model exigem `ALTER TABLE` manual.
- **Performance** → GORM gera queries ligeiramente menos otimizadas que SQL manual em casos extremos. Neste domínio (pagamentos individuais, baixo volume por query), o impacto é irrelevante.
- **Conversão model ↔ domain** → Adiciona um passo de mapeamento em cada operação. Mantido simples com funções `toModel` / `toDomain`.

## Migration Plan

1. Instalar dependências GORM (`go get`)
2. Criar `payment_model.go` com struct e tags
3. Criar `gorm_repository.go` implementando `PaymentRepository`
4. Remover `postgres_repository.go`
5. Atualizar `main.go` para usar `NewGORMRepository`
6. Rodar `go mod tidy` para remover `pgx/v5` direto
7. Build e testes

**Rollback:** reverter os arquivos acima e restaurar `postgres_repository.go` — sem mudança de schema no banco (mesma tabela `payments`).

## Open Questions

*(nenhuma)*
