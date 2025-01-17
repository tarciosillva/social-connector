package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"social-connector/internal/config"
	"social-connector/internal/domain/dto"
	"social-connector/internal/infra/logger"
)

type QueryAIService struct {
	Logger *logger.Logger
}

func NewQueryAIService(logger *logger.Logger) *QueryAIService {
	return &QueryAIService{
		Logger: logger,
	}
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
func (th *QueryAIService) ExecuteQueryAI(queryText string, context string) (dto.QueryAIResponse, error) {
	queryAIHost := config.GetEnv("QUERY_AI_API_HOST")
	if queryAIHost == "" {
		err := "QUERY_AI_API_HOST environment variable not set."
		th.Logger.Error(err)
		return dto.QueryAIResponse{}, fmt.Errorf("%s", err)
	}

	payload := map[string]string{
		"query_text":      queryText,
		"message_context": context,
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
