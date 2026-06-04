package order

type State string

const (
	StateInitialized       State = "initialized"
	StatePaymentAuthorized State = "payment_authorized"
	StateComplete          State = "complete"
)
