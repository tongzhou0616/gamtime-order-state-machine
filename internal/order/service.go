package order

import (
	"fmt"

	"github.com/gametime/order-state-machine/internal/payment"
)

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

func (s *Service) AuthorizePayment(id string) (*Order, error) {
	o, err := s.store.Get(id)
	if err != nil {
		return nil, err
	}

	if o.State != StateInitialized {
		return nil, fmt.Errorf("%w: authorize requires state %q, got %q", ErrInvalidTransition, StateInitialized, o.State)
	}

	if err := s.payment.Authorize(id); err != nil {
		if applyErr := applyTransition(o, StateRejected, fmt.Sprintf("payment declined: %v", err)); applyErr != nil {
			return nil, applyErr
		}
		return s.persist(o)
	}

	if err := applyTransition(o, StatePaymentAuthorized, ""); err != nil {
		return nil, err
	}
	return s.persist(o)
}

func (s *Service) CompleteOrder(id string) (*Order, error) {
	o, err := s.store.Get(id)
	if err != nil {
		return nil, err
	}

	if o.State != StatePaymentAuthorized {
		return nil, fmt.Errorf("%w: complete requires state %q, got %q", ErrInvalidTransition, StatePaymentAuthorized, o.State)
	}

	if err := s.payment.Complete(id); err != nil {
		return nil, fmt.Errorf("complete order: %w", err)
	}

	if err := applyTransition(o, StateComplete, ""); err != nil {
		return nil, err
	}
	return s.persist(o)
}

func (s *Service) persist(o *Order) (*Order, error) {
	if err := s.store.Update(o); err != nil {
		return nil, err
	}
	return o, nil
}
