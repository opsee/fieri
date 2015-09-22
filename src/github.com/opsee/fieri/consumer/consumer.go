package consumer

type Consumer interface {
	Stop() error
}

const (
	Channel = "fieri"
)

type Event struct {
	CustomerId  string `json:"customer_id,omitempty"`
	MessageType string `json:"type"`
	MessageBody []byte `json:"event"`
}
