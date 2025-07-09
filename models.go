package shared

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Order represents an order in the system
type Order struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	UserID      string    `json:"user_id" gorm:"not null"`
	ProductName string    `json:"product_name" gorm:"not null"`
	Quantity    int       `json:"quantity" gorm:"not null"`
	Amount      float64   `json:"amount" gorm:"not null"`
	Status      string    `json:"status" gorm:"default:'pending'"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Payment represents a payment in the system
type Payment struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	OrderID   string    `json:"order_id" gorm:"not null;index"`
	Amount    float64   `json:"amount" gorm:"not null"`
	Status    string    `json:"status" gorm:"default:'pending'"`
	Method    string    `json:"method"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate hooks for generating UUIDs
func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return nil
}

func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// OrderEvent represents an event for pub/sub messaging
type OrderEvent struct {
	EventType string    `json:"event_type"`
	OrderID   string    `json:"order_id"`
	UserID    string    `json:"user_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// PaymentEvent represents a payment event
type PaymentEvent struct {
	EventType string    `json:"event_type"`
	PaymentID string    `json:"payment_id"`
	OrderID   string    `json:"order_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}
