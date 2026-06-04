package order

import (
	"errors"
	"fmt"
	"time"
)

var ErrInvalidTransition = errors.New("invalid state transition")

func recordTransition(o *Order, to State, note string) {
	entry := HistoryEntry{
		From: o.State,
		To:   to,
		At:   time.Now().UTC(),
		Note: note,
	}
	o.State = to
	o.History = append(o.History, entry)
}

func validateTransition(from, to State) error {
	if from.IsTerminal() {
		return fmt.Errorf("%w: order is in terminal state %q", ErrInvalidTransition, from)
	}

	allowed := map[State]map[State]bool{
		StateInitialized: {
			StatePaymentAuthorized: true,
			StateRejected:          true,
		},
	}

	targets, ok := allowed[from]
	if !ok || !targets[to] {
		return fmt.Errorf("%w: cannot transition from %q to %q", ErrInvalidTransition, from, to)
	}
	return nil
}
