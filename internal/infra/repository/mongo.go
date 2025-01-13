package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepository[T any] struct {
	mongo *mongo.Database
}

func NewMongoRepository[T any](mongo *mongo.Database) *MongoRepository[T] {
	return &MongoRepository[T]{mongo: mongo}
}

func (r *MongoRepository[T]) Create(ctx context.Context, collectionName string, entity T) (T, error) {
	collection := r.mongo.Collection(collectionName)
	_, err := collection.InsertOne(ctx, entity)
	return entity, err
}

func (r *MongoRepository[T]) Update(ctx context.Context, collectionName string, conversationalId string, entity T) (T, error) {
	collection := r.mongo.Collection(collectionName)
	filter := bson.M{"conversation_id": conversationalId}

	// Usar UpdateOne com upsert para garantir que o documento seja criado se não existir
	update := bson.M{
		"$set": entity, // Atualizar os campos necessários
	}

	// A opção Upsert garante que um novo documento será criado se o filtro não encontrar nenhum documento
	_, err := collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	return entity, err
}

func (r *MongoRepository[T]) Delete(ctx context.Context, collectionName string, id string) error {
	collection := r.mongo.Collection(collectionName)
	filter := bson.M{"_id": id}
	_, err := collection.DeleteOne(ctx, filter)
	return err
}

func (r *MongoRepository[T]) FindByConversationID(ctx context.Context, collectionName string, conversation_id string) (T, error) {
	var entity T
	collection := r.mongo.Collection(collectionName)
	filter := bson.M{"conversation_id": conversation_id}
	err := collection.FindOne(ctx, filter).Decode(&entity)
	return entity, err
}

func (r *MongoRepository[T]) FindAll(ctx context.Context, collectionName string) ([]T, error) {
	collection := r.mongo.Collection(collectionName)
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entities []T
	for cursor.Next(ctx) {
		var entity T
		if err := cursor.Decode(&entity); err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}
	return entities, nil
}
