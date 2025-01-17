package dto

type InfobipMessagePayload struct {
	From         string         `json:"from"`
	To           string         `json:"to"`
	MessageID    string         `json:"messageId"`
	Content      MessageContent `json:"content"`
	CallbackData string         `json:"callbackData"`
	NotifyURL    string         `json:"notifyUrl"`
	URLOptions   URLOptions     `json:"urlOptions"`
}

type MessageContent struct {
	Text string `json:"text"`
}

type URLOptions struct {
	ShortenURL     bool   `json:"shortenUrl"`
	TrackClicks    bool   `json:"trackClicks"`
	TrackingURL    string `json:"trackingUrl"`
	RemoveProtocol bool   `json:"removeProtocol"`
	CustomDomain   string `json:"customDomain"`
}
