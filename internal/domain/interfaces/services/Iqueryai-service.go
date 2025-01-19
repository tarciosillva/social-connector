package Iservices

import "social-connector/internal/domain/dto"

type IQueryAIService interface {
	ExecuteQueryAI(queryText string, context string) (dto.QueryAIResponse, error)
	ExecuteAudioQueryAI(audioUrl string, audioAuth string, context string) (dto.VoiceQueryAIResponse, error)
}
