package Iservices

import "social-connector/internal/domain/dto"

type IQueryAIService interface {
	ExecuteQueryAI(queryText string, context string) (dto.QueryAIResponse, error)
}
