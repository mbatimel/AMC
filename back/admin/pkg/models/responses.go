package models

import (
	"time"

	"github.com/google/uuid"
)

type LoginResponse struct {
	UserID uuid.UUID `json:"userID"`
	Role   string    `json:"role"`
}

type SessionResponse struct {
	UserID uuid.UUID `json:"userID"`
	Role   string    `json:"role"`
}

type AuditLogEntry struct {
	ID         uuid.UUID `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	ActorLabel string    `json:"actorLabel"`
	Action     string    `json:"action"`
}

type ListAuditLogResponse struct {
	Items []AuditLogEntry `json:"items"`
	Total int             `json:"total"`
}
