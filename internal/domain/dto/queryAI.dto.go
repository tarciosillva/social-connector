package dto

type QueryAIResponse struct {
	Response string   `json:"response"`
	Sources  []string `json:"sources"`
}
