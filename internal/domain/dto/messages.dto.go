package dto

type IWebhookMessage struct {
	Object string         `json:"object"`
	Entry  []WebhookEntry `json:"entry"`
}

type WebhookEntry struct {
	ID      string          `json:"id"`
	Changes []WebhookChange `json:"changes"`
}

type WebhookChange struct {
	Field string       `json:"field"`
	Value WebhookValue `json:"value"`
}

type WebhookValue struct {
	MessagingProduct string               `json:"messaging_product"`
	Metadata         WebhookMetadata      `json:"metadata"`
	Contacts         []WebhookContact     `json:"contacts"`
	Messages         []WebhookMessageData `json:"messages"`
}

type WebhookMetadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

type WebhookContact struct {
	Profile WebhookContactProfile `json:"profile"`
	WaID    string                `json:"wa_id"`
}

type WebhookContactProfile struct {
	Name string `json:"name"`
}

type WebhookMessageData struct {
	From      string      `json:"from"`
	ID        string      `json:"id"`
	Timestamp string      `json:"timestamp"`
	Text      WebhookText `json:"text"`
	Type      string      `json:"type"`
}

type WebhookText struct {
	Body string `json:"body"`
}

type IWhatsAppMessage struct {
	MessagingProduct string              `json:"messaging_product"`
	RecipientType    string              `json:"recipient_type"`
	To               string              `json:"to"`
	Type             string              `json:"type"`
	Text             whatsAppMessageText `json:"text"`
}

type whatsAppMessageText struct {
	PreviewURL bool   `json:"preview_url"`
	Body       string `json:"body"`
}
