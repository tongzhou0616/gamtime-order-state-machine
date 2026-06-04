package payment

import "errors"

var ErrPaymentDeclined = errors.New("payment declined")

type Payment interface {
	Authorize(orderID string) error
	Complete(orderID string) error
	Void(orderID string) error
}
