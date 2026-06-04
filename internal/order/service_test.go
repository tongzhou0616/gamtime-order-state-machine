package order_test

import (
	"errors"
	"strings"
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
	if len(o.History) != 1 {
		t.Fatalf("history length = %d, want 1", len(o.History))
	}
	if o.History[0].To != order.StateRejected {
		t.Fatalf("transition to = %q, want %q", o.History[0].To, order.StateRejected)
	}
}

func TestCompletionFailureVoidSucceeds(t *testing.T) {
	stub := &payment.Stub{
		CompleteFn: func(string) error { return payment.ErrCompletionFailed },
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

	o, err = svc.CompleteOrder(o.ID)
	if err != nil {
		t.Fatalf("CompleteOrder: %v", err)
	}
	if o.State != order.StateCancelled {
		t.Fatalf("state = %q, want %q", o.State, order.StateCancelled)
	}
	if len(o.History) != 2 {
		t.Fatalf("history length = %d, want 2", len(o.History))
	}
	last := o.History[len(o.History)-1]
	if last.To != order.StateCancelled {
		t.Fatalf("last transition to = %q, want %q", last.To, order.StateCancelled)
	}
	if !strings.Contains(last.Note, "complete") {
		t.Fatalf("expected completion failure in note, got %q", last.Note)
	}
}

func TestCompletionFailureVoidFails(t *testing.T) {
	stub := &payment.Stub{
		CompleteFn: func(string) error { return payment.ErrCompletionFailed },
		VoidFn:     func(string) error { return payment.ErrVoidFailed },
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

	o, err = svc.CompleteOrder(o.ID)
	if err != nil {
		t.Fatalf("CompleteOrder: %v", err)
	}
	if o.State != order.StateNeedsAttention {
		t.Fatalf("state = %q, want %q", o.State, order.StateNeedsAttention)
	}
	if o.State == order.StateCancelled {
		t.Fatal("order must not be silently cancelled when void fails")
	}
	last := o.History[len(o.History)-1]
	if !strings.Contains(last.Note, "complete") || !strings.Contains(last.Note, "void") {
		t.Fatalf("note must surface both errors, got %q", last.Note)
	}
}

func TestInvalidTransition(t *testing.T) {
	svc := newTestService(&payment.Stub{})

	o, err := svc.CreateOrder()
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	_, err = svc.CompleteOrder(o.ID)
	if err == nil {
		t.Fatal("expected error completing from initialized")
	}
	if !errors.Is(err, order.ErrInvalidTransition) {
		t.Fatalf("error = %v, want ErrInvalidTransition", err)
	}
}
