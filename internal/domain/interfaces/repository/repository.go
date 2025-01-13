package repository

import "context"

type Repository[T any] interface {
	Create(ctx context.Context, collectionName string, entity T) (T, error)
	Update(ctx context.Context, collectionName string, id string, entity T) (T, error)
	Delete(ctx context.Context, collectionName string, id string) error
	FindByConversationID(ctx context.Context, collectionName string, conversation_id string) (T, error)
	FindAll(ctx context.Context, collectionName string) ([]T, error)
}
