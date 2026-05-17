package database

import (
	"time"

	"github.com/gedanmx/payments-api/internal/core/domain"
)

type paymentModel struct {
	ID                string     `gorm:"column:id;primaryKey"`
	OrderID           string     `gorm:"column:order_id;not null;index"`
	Provider          string     `gorm:"column:provider;not null;default:''"`
	PreferenceID      string     `gorm:"column:preference_id;not null"`
	PaymentID         *string    `gorm:"column:payment_id;index"`
	TransactionAmount float64    `gorm:"column:transaction_amount;not null"`
	NetAmount         *float64   `gorm:"column:net_amount"`
	Currency          string     `gorm:"column:currency;not null"`
	CustomerEmail     string     `gorm:"column:customer_email;not null"`
	CheckoutURL       string     `gorm:"column:checkout_url;not null"`
	ExpiresAt         *time.Time `gorm:"column:expires_at"`
	BusinessStatus    string     `gorm:"column:business_status;not null;default:'PENDING'"`
	SagaStatus        string     `gorm:"column:saga_status;not null;default:'AWAITING_PAYMENT'"`
	Status            string     `gorm:"column:status;not null"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null;autoCreateTime:false"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null;autoUpdateTime:false"`
}

func (paymentModel) TableName() string {
	return "payments"
}

func toModel(p domain.Payment) paymentModel {
	m := paymentModel{
		ID:                p.ID,
		OrderID:           p.OrderID,
		Provider:          p.Provider,
		PreferenceID:      p.PreferenceID,
		TransactionAmount: p.TransactionAmount,
		Currency:          p.Currency,
		CustomerEmail:     p.CustomerEmail,
		CheckoutURL:       p.CheckoutURL,
		ExpiresAt:         p.ExpiresAt,
		BusinessStatus: string(p.BusinessStatus),
		SagaStatus:     string(p.Status),
		Status:         string(p.PaymentStatus),
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
	if p.PaymentID != "" {
		m.PaymentID = &p.PaymentID
	}
	if p.NetAmount != 0 {
		m.NetAmount = &p.NetAmount
	}
	return m
}

func toDomain(m paymentModel) domain.Payment {
	p := domain.Payment{
		ID:                m.ID,
		OrderID:           m.OrderID,
		Provider:          m.Provider,
		PreferenceID:      m.PreferenceID,
		TransactionAmount: m.TransactionAmount,
		Currency:          m.Currency,
		CustomerEmail:     m.CustomerEmail,
		CheckoutURL:       m.CheckoutURL,
		ExpiresAt:         m.ExpiresAt,
		BusinessStatus: domain.BusinessStatus(m.BusinessStatus),
		Status:         domain.SagaStatus(m.SagaStatus),
		PaymentStatus:  domain.PaymentStatus(m.Status),
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
	if m.PaymentID != nil {
		p.PaymentID = *m.PaymentID
	}
	if m.NetAmount != nil {
		p.NetAmount = *m.NetAmount
	}
	return p
}
