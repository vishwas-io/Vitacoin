package store

import (
	"sync"
	"time"
)

// Merchant holds the locally-registered merchant profile.
type Merchant struct {
	ID           string    `json:"merchantId"`
	Address      string    `json:"address"`
	BusinessName string    `json:"businessName"`
	WebhookURL   string    `json:"webhookUrl,omitempty"`
	RegisteredAt time.Time `json:"registeredAt"`
}

// MerchantStore is an in-memory merchant registry.
type MerchantStore struct {
	mu        sync.RWMutex
	byID      map[string]*Merchant
	byAddress map[string]*Merchant
}

var Merchants = &MerchantStore{
	byID:      make(map[string]*Merchant),
	byAddress: make(map[string]*Merchant),
}

func (s *MerchantStore) Save(m *Merchant) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byID[m.ID] = m
	s.byAddress[m.Address] = m
}

func (s *MerchantStore) GetByAddress(address string) (*Merchant, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.byAddress[address]
	return m, ok
}

func (s *MerchantStore) GetByID(id string) (*Merchant, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.byID[id]
	return m, ok
}
