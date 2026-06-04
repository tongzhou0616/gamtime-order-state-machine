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

func (s State) IsTerminal() bool {
	switch s {
	case StateComplete, StateRejected, StateCancelled, StateNeedsAttention:
		return true
	default:
		return false
	}
}

type HistoryEntry struct {
	From State     `json:"from"`
	To   State     `json:"to"`
	At   time.Time `json:"at"`
	Note string    `json:"note,omitempty"`
}

type Order struct {
	ID      string         `json:"id"`
	State   State          `json:"state"`
	History []HistoryEntry `json:"history"`
}
