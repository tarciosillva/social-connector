package entities

import "time"

type UserContext struct {
	ID             string       `json:"_id" bson:"_id"`
	ConversationID string       `json:"conversation_id" bson:"conversation_id"`
	Transcript     []Transcript `json:"transcript" bson:"transcript"`
	Context        string       `json:"context" bson:"context"`
	UpdatedAt      time.Time    `json:"updatedAt"`
}

type Transcript struct {
	Role      string    `json:"role" bson:"role"`
	Message   string    `json:"message" bson:"message"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}
