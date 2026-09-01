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

type CompanyRequestInput struct {
	ContactName string      `json:"contact_name"`
	Email       string      `json:"email"`
	Phone       string      `json:"phone"`
	Company     string      `json:"company"`
	Message     string      `json:"message"`
	Attachments []ImageFile `json:"-"`
}

type CompanyRequestResponse struct {
	Accepted bool `json:"accepted"`
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

type SignupRequest struct {
	ID           uuid.UUID  `json:"id"`
	Company      string     `json:"company"`
	INN          string     `json:"inn"`
	Contact      string     `json:"contact"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	Type         string     `json:"type"`
	Status       string     `json:"status"`
	RejectReason string     `json:"reject_reason"`
	CreatedAt    time.Time  `json:"created_at"`
	DecidedAt    *time.Time `json:"decided_at,omitempty"`
}

type ListSignupRequestsResponse struct {
	Items []SignupRequest `json:"items"`
}

type LegalDoc struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	CurrentVersion string    `json:"current_version"`
	FileURL        string    `json:"file_url"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type LegalDocVersion struct {
	ID        uuid.UUID `json:"id"`
	Version   string    `json:"version"`
	Summary   string    `json:"summary"`
	Author    string    `json:"author"`
	FileURL   string    `json:"file_url"`
	CreatedAt time.Time `json:"created_at"`
}

type ListLegalDocsResponse struct {
	Items []LegalDoc `json:"items"`
}

type ListLegalDocVersionsResponse struct {
	Items []LegalDocVersion `json:"items"`
}

type DeleteLegalDocResponse struct {
	Deleted bool `json:"deleted"`
}

type CertificateInput struct {
	Title     string     `json:"title"`
	SortOrder int        `json:"sort_order"`
	IsActive  bool       `json:"is_active"`
	File      *ImageFile `json:"-"`
}

type Certificate struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
	FileURL   string    `json:"file_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListCertificatesResponse struct {
	Items []Certificate `json:"items"`
}

type DeleteCertificateResponse struct {
	Deleted bool `json:"deleted"`
}
