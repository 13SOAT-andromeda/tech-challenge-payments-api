## 1. Dependências

- [x] 1.1 Adicionar `gorm.io/gorm` e `gorm.io/driver/postgres` com `go get`
- [x] 1.2 Rodar `go mod tidy` para remover `github.com/jackc/pgx/v5` como dependência direta (após remoção do postgres_repository.go)

## 2. Model GORM

- [x] 2.1 Criar `internal/adapters/out/database/payment_model.go` com struct `paymentModel` e tags `gorm:"column:..."`
- [x] 2.2 Mapear campos nullable com ponteiros: `PaymentID *string`, `NetAmount *float64`, `ExpiresAt *time.Time`
- [x] 2.3 Adicionar função `toModel(domain.Payment) paymentModel` no mesmo arquivo
- [x] 2.4 Adicionar função `toDomain(paymentModel) domain.Payment` no mesmo arquivo

## 3. Repositório GORM

- [x] 3.1 Criar `internal/adapters/out/database/gorm_repository.go` com struct `GORMRepository` contendo `*gorm.DB`
- [x] 3.2 Implementar `NewGORMRepository(dsn string) (*GORMRepository, error)` com `gorm.Open`, `db.Ping` e `AutoMigrate`
- [x] 3.3 Implementar `Save(ctx, domain.Payment) error` usando `db.WithContext(ctx).Create(&model)`
- [x] 3.4 Implementar `FindByOrderID(ctx, orderID) (domain.Payment, error)` usando `db.WithContext(ctx).Where(...).First(&model)`
- [x] 3.5 Implementar `FindByPaymentID(ctx, paymentID) (domain.Payment, error)` usando `db.WithContext(ctx).Where(...).First(&model)`
- [x] 3.6 Implementar `UpdatePayment(ctx, orderID, ...) error` usando `db.WithContext(ctx).Model(...).Where(...).Updates(...)` e verificar `RowsAffected`
- [x] 3.7 Tratar `errors.Is(result.Error, gorm.ErrRecordNotFound)` e retornar mensagem "registro não encontrado"

## 4. Remoção do repositório antigo

- [x] 4.1 Deletar `internal/adapters/out/database/postgres_repository.go`

## 5. Atualização do main.go

- [x] 5.1 Substituir `database.NewPostgresRepository` por `database.NewGORMRepository` em `cmd/api/main.go`

## 6. Verificação

- [x] 6.1 Rodar `go build ./...` sem erros
- [x] 6.2 Rodar `go vet ./...` sem warnings
