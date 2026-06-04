package order

import "time"

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
