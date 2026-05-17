package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/gedanmx/payments-api/internal/core/domain"
)

type GORMRepository struct {
	db *gorm.DB
}

func NewGORMRepository(dsn string) (*GORMRepository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("conectar ao banco: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("obter conexão sql: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping ao banco: %w", err)
	}
	if err := db.AutoMigrate(&paymentModel{}); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}
	return &GORMRepository{db: db}, nil
}

func (r *GORMRepository) DB() *gorm.DB { return r.db }

func (r *GORMRepository) Save(ctx context.Context, p domain.Payment) error {
	model := toModel(p)
	result := r.db.WithContext(ctx).Create(&model)
	return result.Error
}

func (r *GORMRepository) FindByOrderID(ctx context.Context, orderID string) (domain.Payment, error) {
	var model paymentModel
	result := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&model)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return domain.Payment{}, fmt.Errorf("registro não encontrado")
	}
	if result.Error != nil {
		return domain.Payment{}, result.Error
	}
	return toDomain(model), nil
}

func (r *GORMRepository) FindByPaymentID(ctx context.Context, paymentID string) (domain.Payment, error) {
	var model paymentModel
	result := r.db.WithContext(ctx).Where("payment_id = ?", paymentID).First(&model)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return domain.Payment{}, fmt.Errorf("registro não encontrado")
	}
	if result.Error != nil {
		return domain.Payment{}, result.Error
	}
	return toDomain(model), nil
}

func (r *GORMRepository) UpdatePayment(ctx context.Context, orderID, paymentID string, netAmount float64, status domain.PaymentStatus, businessStatus domain.BusinessStatus, sagaStatus domain.SagaStatus) error {
	result := r.db.WithContext(ctx).
		Model(&paymentModel{}).
		Where("order_id = ?", orderID).
		Updates(map[string]interface{}{
			"payment_id":      paymentID,
			"net_amount":      netAmount,
			"status":          string(status),
			"business_status": string(businessStatus),
			"saga_status":     string(sagaStatus),
			"updated_at":      time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("pagamento com order_id=%s não encontrado", orderID)
	}
	return nil
}
