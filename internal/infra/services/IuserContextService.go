package services

import (
	"context"
	"fmt"
	"social-connector/internal/domain/entities"
	"social-connector/internal/domain/interfaces/repository"
	repocontants "social-connector/internal/domain/interfaces/repository/contants"
	"social-connector/internal/infra/logger"
)

// UserContextService is the service responsible for UserContext business logic.
type UserContextService struct {
	UserContextRepository repository.Repository[entities.UserContext]
	Ctx                   context.Context
	Logger                *logger.Logger
}

// NewUserContextService creates a new instance of the service.
func NewUserContextService(userContextRepository repository.Repository[entities.UserContext], ctx context.Context, logger *logger.Logger) *UserContextService {
	return &UserContextService{
		UserContextRepository: userContextRepository,
		Ctx:                   ctx,
		Logger:                logger,
	}
}

// Create inserts a new UserContext into the database.
func (ucs *UserContextService) Create(input entities.UserContext) error {
	_, err := ucs.UserContextRepository.Create(ucs.Ctx, repocontants.USER_CONTEXT_COLLECTION, input)
	if err != nil {
		ucs.Logger.Error(fmt.Sprintf("Failed to create UserContext: %v", err))
		return err
	}
	return nil
}

// FindContext retrieves a UserContext by conversationID.
func (ucs *UserContextService) FindContext(conversationID string) (entities.UserContext, error) {
	result, err := ucs.UserContextRepository.FindByConversationID(ucs.Ctx, repocontants.USER_CONTEXT_COLLECTION, conversationID)
	if err != nil {
		ucs.Logger.Error(fmt.Sprintf("Failed to find UserContext with conversationID '%s': %v", conversationID, err))
		return entities.UserContext{}, err
	}

	return result, nil
}

// UpdateUserContext updates a UserContext by ID.
func (ucs *UserContextService) UpdateUserContext(conversationalId string, entity entities.UserContext) (entities.UserContext, error) {
	// Garantir que o campo _id seja removido para evitar conflito
	entity.ID = "" // Definir explicitamente como vazio

	// Agora, chama o repositório para atualizar
	result, err := ucs.UserContextRepository.Update(ucs.Ctx, repocontants.USER_CONTEXT_COLLECTION, conversationalId, entity)
	if err != nil {
		ucs.Logger.Error(fmt.Sprintf("Failed to update UserContext with conversationalId '%s': %v", conversationalId, err))
		return entities.UserContext{}, err
	}

	return result, nil
}
