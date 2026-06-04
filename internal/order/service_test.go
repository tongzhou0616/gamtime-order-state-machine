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
