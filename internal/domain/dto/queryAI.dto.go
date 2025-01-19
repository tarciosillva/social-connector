package dto

type QueryAIResponse struct {
	Response string   `json:"response"`
	Sources  []string `json:"sources"`
}

type VoiceQueryAIResponse struct {
	Response  string `json:"response"`
	AudioLink string `json:"audio_link"`
	QueryText string `json:"query_text"`
}
