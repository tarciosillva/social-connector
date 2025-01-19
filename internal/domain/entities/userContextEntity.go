package entities

import "time"

type UserContext struct {
	ConversationID string       `json:"conversation_id" bson:"conversation_id"`
	Transcript     []Transcript `json:"transcript" bson:"transcript"`
	Context        string       `json:"context" bson:"context"`
	UpdatedAt      time.Time    `json:"updatedAt"`
}

type Transcript struct {
	Role      string    `json:"role" bson:"role"`
	Message   string    `json:"message" bson:"message"`
	Audio     string    `json:"audio" bson:"audio,omitempty"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}
