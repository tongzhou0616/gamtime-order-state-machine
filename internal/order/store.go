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
	s.orders[id] = o
	return o, nil
}
