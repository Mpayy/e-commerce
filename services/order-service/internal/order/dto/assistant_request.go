package dto

type AssistantRequest struct {
	Message        string `json:"message" validate:"required"`
	ConversationID string `json:"conversation_id" validate:"omitempty"`
}
