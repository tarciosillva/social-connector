package provider

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
	"social-connector/internal/infra/logger"
	"strings"
)

type InfobipWhatsAppProvider struct {
	Logger     *logger.Logger
	HttpClient *http.Client
}

func NewInfobipWhatsAppProvider(logger *logger.Logger, httpClient *http.Client) *InfobipWhatsAppProvider {
	return &InfobipWhatsAppProvider{Logger: logger, HttpClient: httpClient}
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
func (th *InfobipWhatsAppProvider) SendTextMessage(to, message string) error {
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

	authToken, err := th.GenerateOAuth2Token()
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

	fmt.Println(authToken)

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

// sendAudioMessage sends a audio message to a recipient's phone number using the Infobip API.
//
// Parameters:
//   - to: string - The recipient's phone number in international format (including the country code).
//   - audio_url: string - The content of the audio message to be sent.
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
func (th *InfobipWhatsAppProvider) SendAudioMessage(to, audioLink string) error {
	if to == "" || audioLink == "" {
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

	authToken, err := th.GenerateOAuth2Token()
	if err != nil {
		return err
	}

	payloadData := struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Content struct {
			MediaUrl string `json:"mediaUrl"`
		} `json:"content"`
	}{
		From: from,
		To:   to,
	}
	payloadData.Content.MediaUrl = audioLink

	payload, err := json.Marshal(payloadData)
	if err != nil {
		th.Logger.Error(fmt.Sprintf("Failed to marshal payload %v", err))
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/whatsapp/1/message/audio", infoBipUrl)
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

func (th *InfobipWhatsAppProvider) GenerateOAuth2Token() (*dto.TokenResponse, error) {
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
		return &dto.TokenResponse{}, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	resp, err := th.HttpClient.Do(req)
	if err != nil {
		return &dto.TokenResponse{}, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return &dto.TokenResponse{}, fmt.Errorf("unexpected HTTP status: %d, response: %s", resp.StatusCode, string(body))
	}

	var tokenResponse dto.TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		return &dto.TokenResponse{}, fmt.Errorf("error decoding response JSON: %v", err)
	}

	return &tokenResponse, nil
}
