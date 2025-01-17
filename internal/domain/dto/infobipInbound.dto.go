package dto

type InboundResponse struct {
	Results             []Result `json:"results"`
	MessageCount        int      `json:"messageCount"`
	PendingMessageCount int      `json:"pendingMessageCount"`
}

type Result struct {
	From            string  `json:"from"`
	To              string  `json:"to"`
	IntegrationType string  `json:"integrationType"`
	ReceivedAt      string  `json:"receivedAt"`
	MessageID       string  `json:"messageId"`
	PairedMessageID *string `json:"pairedMessageId"`
	CallbackData    *string `json:"callbackData"`
	Message         Message `json:"message"`
	Contact         Contact `json:"contact"`
	Price           Price   `json:"price"`
}

type Message struct {
	Type string `json:"type"`
	Text string `json:"text"`
	Url  string `json:"url"`
}

type Contact struct {
	Name string `json:"name"`
}

type Price struct {
	PricePerMessage float64 `json:"pricePerMessage"`
	Currency        string  `json:"currency"`
}
