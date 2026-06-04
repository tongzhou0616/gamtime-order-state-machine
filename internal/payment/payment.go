package payment

import "errors"

var (
	ErrPaymentDeclined  = errors.New("payment declined")
	ErrCompletionFailed = errors.New("completion failed")
	ErrVoidFailed       = errors.New("void failed")
)

type Payment interface {
	Authorize(orderID string) error
	Complete(orderID string) error
	Void(orderID string) error
}

type Stub struct {
	AuthorizeFn func(string) error
	CompleteFn  func(string) error
	VoidFn      func(string) error
}

func (s *Stub) Authorize(orderID string) error {
	if s.AuthorizeFn != nil {
		return s.AuthorizeFn(orderID)
	}
	return nil
}

func (s *Stub) Complete(orderID string) error {
	if s.CompleteFn != nil {
		return s.CompleteFn(orderID)
	}
	return nil
}

func (s *Stub) Void(orderID string) error {
	if s.VoidFn != nil {
		return s.VoidFn(orderID)
	}
	return nil
}
