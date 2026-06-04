package order

import (
	"errors"
	"fmt"
	"sync"
)

var ErrOrderNotFound = errors.New("order not found")

type Store interface {
	Create() (*Order, error)
	Get(id string) (*Order, error)
	Update(o *Order) error
}

type MemoryStore struct {
	mu     sync.RWMutex
	orders map[string]*Order
	nextID int
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		orders: make(map[string]*Order),
	}
}

func (s *MemoryStore) Create() (*Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	id := fmt.Sprintf("ord_%d", s.nextID)
	o := &Order{
		ID:    id,
		State: StateInitialized,
	}
	s.orders[id] = cloneOrder(o)
	return cloneOrder(o), nil
}

func (s *MemoryStore) Get(id string) (*Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	o, ok := s.orders[id]
	if !ok {
		return nil, ErrOrderNotFound
	}
	return cloneOrder(o), nil
}

func (s *MemoryStore) Update(o *Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.orders[o.ID]; !ok {
		return ErrOrderNotFound
	}
	s.orders[o.ID] = cloneOrder(o)
	return nil
}

func cloneOrder(o *Order) *Order {
	history := make([]HistoryEntry, len(o.History))
	copy(history, o.History)
	return &Order{
		ID:      o.ID,
		State:   o.State,
		History: history,
	}
}
