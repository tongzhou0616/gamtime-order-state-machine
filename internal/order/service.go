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
	return &Service{
		store:   store,
		payment: p,
	}
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
		if err := s.store.Update(o); err != nil {
			return nil, err
		}
		return o, nil
	}

	if err := applyTransition(o, StatePaymentAuthorized, ""); err != nil {
		return nil, err
	}
	if err := s.store.Update(o); err != nil {
		return nil, err
	}
	return o, nil
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
		return s.handleCompletionFailure(o, err)
	}

	if err := applyTransition(o, StateComplete, ""); err != nil {
		return nil, err
	}
	if err := s.store.Update(o); err != nil {
		return nil, err
	}
	return o, nil
}

func (s *Service) handleCompletionFailure(o *Order, completeErr error) (*Order, error) {
	voidErr := s.payment.Void(o.ID)
	if voidErr != nil {
		note := fmt.Sprintf("complete: %v; void: %v", completeErr, voidErr)
		if err := applyTransition(o, StateNeedsAttention, note); err != nil {
			return nil, err
		}
	} else {
		note := fmt.Sprintf("complete: %v", completeErr)
		if err := applyTransition(o, StateCancelled, note); err != nil {
			return nil, err
		}
	}

	if err := s.store.Update(o); err != nil {
		return nil, err
	}
	return o, nil
}
