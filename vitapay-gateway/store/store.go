package store

import (
	"sync"
	"time"
)

// PaymentStatus represents the lifecycle state of a payment.
type PaymentStatus string

const (
	StatusPending   PaymentStatus = "pending"
	StatusConfirmed PaymentStatus = "confirmed"
	StatusExpired   PaymentStatus = "expired"
	StatusFailed    PaymentStatus = "failed"
)

// Payment is the core payment record.
type Payment struct {
	ID              string        `json:"paymentId"`
	MerchantAddress string        `json:"merchantAddress"`
	Amount          string        `json:"amount"`
	Denom           string        `json:"denom"`
	Memo            string        `json:"memo"`
	Status          PaymentStatus `json:"status"`
	CreatedAt       time.Time     `json:"createdAt"`
	ExpiresAt       time.Time     `json:"expiresAt"`
	ConfirmedAt     *time.Time    `json:"confirmedAt,omitempty"`
	TxHash          string        `json:"txHash,omitempty"`
	WebhookURL      string        `json:"webhookUrl,omitempty"`
	QRData          string        `json:"qrData"`
	DeepLink        string        `json:"deepLink"`
}

// MemStore is an in-memory payment store (swap out for Supabase when DATABASE_URL is set).
type MemStore struct {
	mu       sync.RWMutex
	payments map[string]*Payment
}

var Default = &MemStore{payments: make(map[string]*Payment)}

func (s *MemStore) Save(p *Payment) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.payments[p.ID] = p
}

func (s *MemStore) Get(id string) (*Payment, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.payments[id]
	return p, ok
}

func (s *MemStore) Update(id string, fn func(*Payment)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.payments[id]
	if !ok {
		return false
	}
	fn(p)
	return true
}

func (s *MemStore) ListByMerchant(address string) []*Payment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*Payment
	for _, p := range s.payments {
		if p.MerchantAddress == address {
			out = append(out, p)
		}
	}
	return out
}
