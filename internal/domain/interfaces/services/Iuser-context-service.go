package Iservices

import "social-connector/internal/domain/entities"

// UserContextServiceInterface defines the methods the service must implement.
type IUserContextService interface {
	Create(input entities.UserContext) error
	FindContext(conversationID string) (entities.UserContext, error)
	UpdateUserContext(conversationalId string, entity entities.UserContext) (entities.UserContext, error)
}
