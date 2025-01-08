package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"social-connector/internal/config"
	"social-connector/internal/domain/dto"
	"social-connector/internal/infra/logger"
	"social-connector/internal/util"
)

type HttpHandlers struct {
	Logger      *logger.Logger
	VerifyToken string
}

func NewHttpHandlers(logger *logger.Logger, verifyToken string) *HttpHandlers {
	return &HttpHandlers{Logger: logger, VerifyToken: verifyToken}
}

// Webhook is a unified handler for WhatsApp webhook requests.
//
// This function handles both verification requests (GET) and event notifications (POST)
// sent by the WhatsApp Meta API to the configured webhook URL. It delegates the actual
// handling to specific methods (`handleVerification` for GET and `handleWebhookEvent` for POST).
//
// Parameters:
// - w (http.ResponseWriter): The HTTP response writer used to send a response back to the client.
// - r (*http.Request): The HTTP request object containing details about the incoming request.
//
// Functionality:
//   - If the request method is GET, the function calls `handleVerification` to handle the webhook
//     verification process.
//   - If the request method is POST, the function calls `handleWebhookEvent` to process incoming
//     webhook events, such as messages or status updates.
//   - For any other HTTP methods, the function responds with a 405 Method Not Allowed error.
//
// HTTP Status Codes:
// - 200 OK: Returned by `handleVerification` or `handleWebhookEvent` upon successful processing.
// - 403 Forbidden: Returned by `handleVerification` if the verification token is invalid.
// - 405 Method Not Allowed: Returned for HTTP methods other than GET or POST.
func (th *HttpHandlers) Webhook(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		th.handleVerification(w, r)
		return
	}

	if r.Method == http.MethodPost {
		th.handleWebhookEvent(w, r)
		return
	}

	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
}

// handleVerification handles the webhook verification process for WhatsApp Meta API (GET request).
//
// This function is used by WhatsApp to verify the webhook endpoint during setup.
// When WhatsApp sends a GET request to the webhook URL, it includes specific query parameters
// (like `hub.mode`, `hub.challenge`, and `hub.verify_token`) to validate the endpoint.
//
// Parameters:
// - w (http.ResponseWriter): The HTTP response writer used to send a response back to WhatsApp.
// - r (*http.Request): The HTTP request object containing the query parameters for verification.
//
// Expected Query Parameters:
// - hub.mode (string): Should be equal to "subscribe".
// - hub.challenge (string): A random string sent by WhatsApp that should be echoed back in the response.
// - hub.verify_token (string): A user-defined token that matches the one configured in the WhatsApp App Dashboard.
//
// Response:
//   - If the `hub.verify_token` matches the token you configured, the function responds with the `hub.challenge` value
//     and a 200 status code to confirm the webhook verification.
//   - If the token doesn't match, the function responds with a 403 status code (Forbidden).
func (th *HttpHandlers) handleVerification(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	mode := query.Get("hub.mode")
	token := query.Get("hub.verify_token")
	challenge := query.Get("hub.challenge")

	if mode == "subscribe" && token == th.VerifyToken {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
		return
	}

	http.Error(w, "Forbidden", http.StatusForbidden)
}

// handleWebhookEvent handles incoming webhook events from the WhatsApp Meta API (POST request).
//
// This function processes various event notifications sent by WhatsApp to the configured webhook URL.
// These events may include message notifications, message status updates (e.g., sent, delivered, read),
// and other relevant webhook payloads.
//
// Parameters:
// - w (http.ResponseWriter): The HTTP response writer used to send a response back to WhatsApp.
// - r (*http.Request): The HTTP request object containing the event data in the request body.
//
// Response:
// - Respond with HTTP status 200 and an empty body to acknowledge receipt of the event.
// - If an error occurs during processing, respond with an appropriate HTTP status code.
func (th *HttpHandlers) handleWebhookEvent(w http.ResponseWriter, r *http.Request) {
	var body dto.IWebhookMessage

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		th.Logger.Error(fmt.Sprintf("Invalid JSON payload: %s", err.Error()))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(body.Entry) == 0 {
		th.Logger.Warn("Received webhook event with no entries.")
		http.Error(w, "No entries found in the webhook event", http.StatusBadRequest)
		return
	}

	lastEntry := body.Entry[len(body.Entry)-1]
	lastChange := lastEntry.Changes[len(lastEntry.Changes)-1]
	lastMessage := lastChange.Value.Messages[len(lastChange.Value.Messages)-1]

	from := lastMessage.From
	text := lastMessage.Text.Body

	result, err := th.executeQueryAI(text)
	if err != nil {
		return
	}

	to := util.AddNineToPhoneNumber(from)

	th.sendWhatsAppMessage(to, result.Response)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("EVENT_RECEIVED"))
}

// executeQueryAI processes a query text using an AI service and returns the response.
//
// Parameters:
// - queryText (string): The input text query to be processed by the AI service.
//
// Returns:
//   - dto.QueryAIResponse: A structured response object containing the AI's output,
//     which may include answers, insights, or any relevant data returned by the AI.
//   - error: Returns an error if the query processing fails or if there is an issue
//     with the integration to the AI service. Returns nil if the query is processed successfully.
//
// Note:
// This function depends on an AI service integration, such as OpenAI, Google Cloud AI,
// or another machine learning model API.
func (th *HttpHandlers) executeQueryAI(queryText string) (dto.QueryAIResponse, error) {
	queryAIHost := config.GetEnv("QUERY_AI_API_HOST")
	if queryAIHost == "" {
		err := "QUERY_AI_API_HOST environment variable not set."
		th.Logger.Error(err)
		return dto.QueryAIResponse{}, fmt.Errorf("%s", err)
	}

	payload := map[string]string{
		"query_text": queryText,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		th.Logger.Error(fmt.Sprintf("Failed to marshal payload: %s", err.Error()))
		return dto.QueryAIResponse{}, err
	}

	resp, err := http.Post(queryAIHost+"/query", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		th.Logger.Error(fmt.Sprintf("Failed to send POST request: %s", err.Error()))
		return dto.QueryAIResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		th.Logger.Error(fmt.Sprintf("Failed to read response body: %s", err.Error()))
		return dto.QueryAIResponse{}, err
	}

	var queryResponse dto.QueryAIResponse
	if err := json.Unmarshal(body, &queryResponse); err != nil {
		th.Logger.Error(fmt.Sprintf("Failed to unmarshal response body: %s", err.Error()))
		return dto.QueryAIResponse{}, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return queryResponse, nil
}

// sendWhatsAppMessage sends a WhatsApp message to a specified recipient.
//
// Parameters:
//   - to (string): The recipient's phone number in international format,
//     including the country code. Example: "+5511999998888".
//   - messageBody (string): The content of the message to be sent.
//
// Returns:
//   - error: Returns an error if there is an issue during the message sending process,
//     or nil if the message is sent successfully.
//
// Note:
// This function relies on the implementation or integration with a WhatsApp message-sending service,
// such as Twilio, Meta API, or another similar API. Ensure that the service is properly configured
// before using this function.
func (th *HttpHandlers) sendWhatsAppMessage(to string, messageBody string) error {
	graphAPIURL := config.GetEnv("GRAPH_API_URL")
	version := config.GetEnv("GRAPH_API_VERSION")
	accessToken := config.GetEnv("WHATSAPP_ACCESS_TOKEN")
	phoneNumberID := config.GetEnv("WHATSAPP_PHONE_NUMBER_ID")

	if graphAPIURL == "" || version == "" || phoneNumberID == "" || accessToken == "" {
		th.Logger.Error("One or more required environment variables are missing.")
		return fmt.Errorf("missing environment variables")
	}

	apiURL := fmt.Sprintf("%s/%s/%s/messages", graphAPIURL, version, phoneNumberID)

	message := dto.IWhatsAppMessage{
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               to,
		Type:             "text",
	}
	message.Text.PreviewURL = false
	message.Text.Body = messageBody

	payloadBytes, err := json.Marshal(message)
	if err != nil {
		th.Logger.Error(fmt.Sprintf("Failed to marshal payload: %s", err.Error()))
		return err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		th.Logger.Error(fmt.Sprintf("Failed to create HTTP request: %s", err.Error()))
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		th.Logger.Error(fmt.Sprintf("Failed to send HTTP request: %s", err.Error()))
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		th.Logger.Error(fmt.Sprintf("Failed to read response body: %s", err.Error()))
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		th.Logger.Error(fmt.Sprintf("API returned an error. Status: %d, Body: %s", resp.StatusCode, string(body)))
		return fmt.Errorf("API returned error: %s", string(body))
	}

	th.Logger.Info(fmt.Sprintf("WhatsApp message sent successfully. Response: %s", string(body)))
	return nil
}
