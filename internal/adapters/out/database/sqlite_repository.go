package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/gedanmx/payments-api/internal/core/domain"
)

const schema = `
CREATE TABLE IF NOT EXISTS payments (
	id                 TEXT PRIMARY KEY,
	order_id           TEXT NOT NULL,
	correlation_id     TEXT NOT NULL DEFAULT '',
	provider           TEXT NOT NULL DEFAULT '',
	preference_id      TEXT NOT NULL,
	payment_id         TEXT,
	transaction_amount REAL NOT NULL,
	net_amount         REAL,
	currency           TEXT NOT NULL,
	customer_email     TEXT NOT NULL,
	checkout_url       TEXT NOT NULL,
	expires_at         DATETIME,
	business_status    TEXT NOT NULL DEFAULT 'PENDING',
	saga_status        TEXT NOT NULL DEFAULT 'AWAITING_PAYMENT',
	status             TEXT NOT NULL,
	created_at         DATETIME NOT NULL,
	updated_at         DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_payments_order_id   ON payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_payment_id ON payments(payment_id);
`

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dsn string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("abrir banco: %w", err)
	}
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("criar schema: %w", err)
	}
	return &SQLiteRepository{db: db}, nil
}

func (r *SQLiteRepository) Save(ctx context.Context, p domain.Payment) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO payments
			(id, order_id, correlation_id, provider, preference_id, payment_id,
			 transaction_amount, net_amount, currency, customer_email, checkout_url,
			 expires_at, business_status, saga_status, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.OrderID, p.CorrelationID, p.Provider, p.PreferenceID,
		nullableStr(p.PaymentID), p.TransactionAmount, nullableFloat(p.NetAmount),
		p.Currency, p.CustomerEmail, p.CheckoutURL,
		nullableTime(p.ExpiresAt),
		string(p.BusinessStatus), string(p.SagaStatus), string(p.Status),
		p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *SQLiteRepository) FindByOrderID(ctx context.Context, orderID string) (domain.Payment, error) {
	return r.scanOne(ctx, `SELECT * FROM payments WHERE order_id = ? LIMIT 1`, orderID)
}

func (r *SQLiteRepository) FindByPaymentID(ctx context.Context, paymentID string) (domain.Payment, error) {
	return r.scanOne(ctx, `SELECT * FROM payments WHERE payment_id = ? LIMIT 1`, paymentID)
}

func (r *SQLiteRepository) UpdatePayment(ctx context.Context, orderID, paymentID string, netAmount float64, status domain.PaymentStatus, businessStatus domain.BusinessStatus, sagaStatus domain.SagaStatus) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE payments
		 SET payment_id = ?, net_amount = ?, status = ?, business_status = ?, saga_status = ?, updated_at = ?
		 WHERE order_id = ?`,
		paymentID, netAmount, string(status), string(businessStatus), string(sagaStatus), time.Now(), orderID,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("pagamento com order_id=%s não encontrado", orderID)
	}
	return nil
}

func (r *SQLiteRepository) scanOne(ctx context.Context, query string, arg interface{}) (domain.Payment, error) {
	row := r.db.QueryRowContext(ctx, query, arg)
	var p domain.Payment
	var paymentID sql.NullString
	var netAmount sql.NullFloat64
	var expiresAt sql.NullTime
	var status, businessStatus, sagaStatus string

	err := row.Scan(
		&p.ID, &p.OrderID, &p.CorrelationID, &p.Provider, &p.PreferenceID, &paymentID,
		&p.TransactionAmount, &netAmount, &p.Currency, &p.CustomerEmail,
		&p.CheckoutURL, &expiresAt,
		&businessStatus, &sagaStatus, &status,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return domain.Payment{}, fmt.Errorf("registro não encontrado")
	}
	if err != nil {
		return domain.Payment{}, err
	}

	p.PaymentID = paymentID.String
	p.NetAmount = netAmount.Float64
	p.Status = domain.PaymentStatus(status)
	p.BusinessStatus = domain.BusinessStatus(businessStatus)
	p.SagaStatus = domain.SagaStatus(sagaStatus)
	if expiresAt.Valid {
		t := expiresAt.Time
		p.ExpiresAt = &t
	}
	return p, nil
}

func nullableStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullableFloat(f float64) interface{} {
	if f == 0 {
		return nil
	}
	return f
}

func nullableTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}
