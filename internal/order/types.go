package order

import "time"

type State string

const (
	StateInitialized        State = "initialized"
	StatePaymentAuthorized  State = "payment_authorized"
	StateComplete           State = "complete"
	StateRejected           State = "rejected"
	StateCancelled          State = "cancelled"
	StateNeedsAttention     State = "needs_attention"
)

type HistoryEntry struct {
	From State     `json:"from"`
	To   State     `json:"to"`
	At   time.Time `json:"at"`
	Note string    `json:"note,omitempty"`
}
