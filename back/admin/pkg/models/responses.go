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

type ImageFile struct {
	FileName string `json:"fileName"`
	Content  []byte `json:"-"`
}

type BannerInput struct {
	Title     string     `json:"title"`
	Subtitle  string     `json:"subtitle"`
	Link      string     `json:"link"`
	SortOrder int        `json:"sort_order"`
	IsActive  bool       `json:"is_active"`
	DateFrom  string     `json:"dateFrom"`
	DateTo    string     `json:"dateTo"`
	Image     *ImageFile `json:"-"`
}

type Banner struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Subtitle  string    `json:"subtitle"`
	Image     string    `json:"image"`
	Link      string    `json:"link"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
	DateFrom  string    `json:"dateFrom"`
	DateTo    string    `json:"dateTo"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BannersSettings struct {
	DelaySec int      `json:"delay_sec"`
	Items    []Banner `json:"items"`
}

type DeleteBannerResponse struct {
	Deleted bool `json:"deleted"`
}
