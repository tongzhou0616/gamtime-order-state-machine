package order_test

import (
	"testing"

	"github.com/gametime/order-state-machine/internal/order"
	"github.com/gametime/order-state-machine/internal/payment"
)

func newTestService(stub *payment.Stub) *order.Service {
	return order.NewService(order.NewMemoryStore(), stub)
}

func TestHappyPath(t *testing.T) {
	svc := newTestService(&payment.Stub{})

	o, err := svc.CreateOrder()
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	if o.State != order.StateInitialized {
		t.Fatalf("initial state = %q, want %q", o.State, order.StateInitialized)
	}

	o, err = svc.AuthorizePayment(o.ID)
	if err != nil {
		t.Fatalf("AuthorizePayment: %v", err)
	}
	if o.State != order.StatePaymentAuthorized {
		t.Fatalf("state after authorize = %q, want %q", o.State, order.StatePaymentAuthorized)
	}

	o, err = svc.CompleteOrder(o.ID)
	if err != nil {
		t.Fatalf("CompleteOrder: %v", err)
	}
	if o.State != order.StateComplete {
		t.Fatalf("final state = %q, want %q", o.State, order.StateComplete)
	}
	if len(o.History) != 2 {
		t.Fatalf("history length = %d, want 2", len(o.History))
	}
	for _, entry := range o.History {
		if entry.At.IsZero() {
			t.Fatal("history entry missing timestamp")
		}
	}
}

func TestPaymentDecline(t *testing.T) {
	voidCalls := 0
	stub := &payment.Stub{
		AuthorizeFn: func(string) error { return payment.ErrPaymentDeclined },
		VoidFn: func(string) error {
			voidCalls++
			return nil
		},
	}
	svc := newTestService(stub)

	o, err := svc.CreateOrder()
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	o, err = svc.AuthorizePayment(o.ID)
	if err != nil {
		t.Fatalf("AuthorizePayment: %v", err)
	}
	if o.State != order.StateRejected {
		t.Fatalf("state = %q, want %q", o.State, order.StateRejected)
	}
	if voidCalls != 0 {
		t.Fatalf("void called %d times, want 0", voidCalls)
	}
}

func TestCompletionFailureVoidSucceeds(t *testing.T) {
	stub := &payment.Stub{
		CompleteFn: func(string) error { return payment.ErrCompletionFailed },
	}
	svc := newTestService(stub)

	o, _ := svc.CreateOrder()
	o, _ = svc.AuthorizePayment(o.ID)
	o, err := svc.CompleteOrder(o.ID)
	if err != nil {
		t.Fatalf("CompleteOrder: %v", err)
	}
	if o.State != order.StateCancelled {
		t.Fatalf("state = %q, want cancelled", o.State)
	}
}
