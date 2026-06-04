package order

import "github.com/gametime/order-state-machine/internal/payment"

type Service struct {
	store   Store
	payment payment.Payment
}

func NewService(store Store, p payment.Payment) *Service {
	return &Service{store: store, payment: p}
}

func (s *Service) CreateOrder() (*Order, error) {
	return s.store.Create()
}

func (s *Service) GetOrder(id string) (*Order, error) {
	return s.store.Get(id)
}
