package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"social-connector/internal/config"
	"social-connector/internal/domain/dto"
	"social-connector/internal/domain/entities"
	Iservices "social-connector/internal/domain/interfaces/services"
	"social-connector/internal/infra/logger"
	"strings"
	"time"
)

type InfobipHandlers struct {
	Logger             *logger.Logger
	UserContextService Iservices.IUserContextService
	QueryAIService     Iservices.IQueryAIService
	HttpClient         *http.Client
}

func NewInfobipHandlers(logger *logger.Logger, userContextService Iservices.IUserContextService, queryAIService Iservices.IQueryAIService, httpClient *http.Client) *InfobipHandlers {
	return &InfobipHandlers{Logger: logger, UserContextService: userContextService, QueryAIService: queryAIService, HttpClient: httpClient}
}

func (th *InfobipHandlers) InfoBipWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var webhookRequest dto.InboundResponse
	err := json.NewDecoder(r.Body).Decode(&webhookRequest)
	if err != nil {
		http.Error(w, "Erro ao processar o JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	lastMessage := webhookRequest.Results[webhookRequest.MessageCount-1].Message.Text
	from := webhookRequest.Results[webhookRequest.MessageCount-1].From
	conversationalId := from

	go func() {
		defer func() {
			if r := recover(); r != nil {
				th.Logger.Error(fmt.Sprintf("Recovered from panic: %v", r))
			}
		}()

		userContext, err := th.UserContextService.FindContext(conversationalId)
		if err != nil {
			th.Logger.Info(fmt.Sprintf("Context not found for conversation ID %s. Initializing new context.", conversationalId))
			userContext = entities.UserContext{
				ConversationID: conversationalId,
				Transcript:     []entities.Transcript{},
				Context:        "",
			}

			err := th.UserContextService.Create(userContext)
			if err != nil {
				th.Logger.Error(fmt.Sprintf("Error to create a new context to %s. Err: %v", conversationalId, err))
			}
		}

		userContext.Transcript = append(userContext.Transcript, entities.Transcript{
			Role:      "user",
			Message:   lastMessage,
			Timestamp: time.Now(),
		})

		result, err := th.QueryAIService.ExecuteQueryAI(lastMessage, userContext.Context)
		if err != nil {
			th.Logger.Error(fmt.Sprintf("Failed to execute AI query: %s", err.Error()))
			return
		}

		userContext.Transcript = append(userContext.Transcript, entities.Transcript{
			Role:      "agent",
			Message:   result.Response,
			Timestamp: time.Now(),
		})

		if len(userContext.Transcript) > 1 {
			lastAgentMessage := userContext.Transcript[len(userContext.Transcript)-1].Message
			userContext.Context = lastAgentMessage
		} else {
			userContext.Context = result.Response
		}

		userContext.UpdatedAt = time.Now()
		if _, err := th.UserContextService.UpdateUserContext(conversationalId, userContext); err != nil {
			th.Logger.Error(fmt.Sprintf("Failed to update user context: %s", err.Error()))
			return
		}

		messagesSplit := strings.Split(result.Response, ".")

		th.Logger.Info(fmt.Sprintf("Sending AI response messages to WhatsApp number: %s", from))
		for i, message := range messagesSplit {
			if strings.TrimSpace(message) != "" {
				if i > 0 {
					time.Sleep(2 * time.Second)
				}
				if err := th.SendTextMessage(from, message); err != nil {
					th.Logger.Error(fmt.Sprintf("Failed to send WhatsApp message to %s: %s", from, err.Error()))
					return
				}
			}
		}
	}()

	w.WriteHeader(http.StatusOK)
}

// sendTextMessage sends a text message to a recipient's phone number using the Infobip API.
//
// Parameters:
//   - to: string - The recipient's phone number in international format (including the country code).
//   - message: string - The content of the text message to be sent.
//
// Returns:
//   - error: Returns an error if any step of the process fails, including input validation,
//     payload construction, HTTP request failure, or unexpected API response.
//
// Dependencies:
//   - Environment variables:
//   - INFOBIP_URL: The base URL of the Infobip API.
//   - AUTH_TOKEN: The authentication token required to authorize the request.
//   - WHATSAPP_PHONE_NUMBER: The registered phone number used to send messages via Infobip.
//   - Data structures: This function relies on the `dto.InfobipMessagePayload` type to create the message payload.
func (th *InfobipHandlers) SendTextMessage(to, message string) error {
	if to == "" || message == "" {
		return fmt.Errorf("recipient (to) and message cannot be empty")
	}

	requiredConfigs := []struct {
		name  string
		value string
	}{
		{"INFOBIP_URL", config.GetEnv("INFOBIP_URL")},
		{"WHATSAPP_PHONE_NUMBER", config.GetEnv("WHATSAPP_PHONE_NUMBER")},
	}

	for _, configItem := range requiredConfigs {
		if configItem.value == "" {
			th.Logger.Error(fmt.Sprintf("%s is not set", configItem.name))
			return fmt.Errorf("%s is not set", configItem.name)
		}
	}

	infoBipUrl := requiredConfigs[0].value
	from := requiredConfigs[1].value

	authToken, err := th.generateOAuth2Token()
	if err != nil {
		return err
	}

	payloadData := struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Content struct {
			Text string `json:"text"`
		} `json:"content"`
	}{
		From: from,
		To:   to,
	}
	payloadData.Content.Text = message

	payload, err := json.Marshal(payloadData)
	if err != nil {
		th.Logger.Error(fmt.Sprintf("Failed to marshal payload %v", err))
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/whatsapp/1/message/text", infoBipUrl)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(payload)))
	if err != nil {
		th.Logger.Error(fmt.Sprintf("Failed to create HTTP request %v", err))
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken.AccessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := th.HttpClient.Do(req)
	if err != nil {
		th.Logger.Error(fmt.Sprintf("HTTP request failed %v", err))
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(res.Body)
		th.Logger.Error(fmt.Sprintf("Unexpected HTTP status %s response_body %s", res.Status, string(body)))
		return fmt.Errorf("unexpected HTTP status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		th.Logger.Error(fmt.Sprintf("Failed to read response body %v", err))
		return fmt.Errorf("failed to read response body: %w", err)
	}

	th.Logger.Info(fmt.Sprintf("Message sent successfully %s response_body %s", res.Status, string(body)))
	return nil
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (th *InfobipHandlers) generateOAuth2Token() (*TokenResponse, error) {
	infobipUrl := config.GetEnv("INFOBIP_URL")
	apiURL := fmt.Sprintf("%s/auth/1/oauth2/token", infobipUrl)

	clientID := os.Getenv("INFOBIP_CLIENT_ID")
	clientSecret := os.Getenv("INFOBIP_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		log.Fatal("Environment variables INFOBIP_CLIENT_ID and INFOBIP_CLIENT_SECRET must be set.")
	}

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return &TokenResponse{}, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	resp, err := th.HttpClient.Do(req)
	if err != nil {
		return &TokenResponse{}, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return &TokenResponse{}, fmt.Errorf("unexpected HTTP status: %d, response: %s", resp.StatusCode, string(body))
	}

	var tokenResponse TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		return &TokenResponse{}, fmt.Errorf("error decoding response JSON: %v", err)
	}

	return &tokenResponse, nil
}
