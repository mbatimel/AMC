package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	UserStatusPending = "pending"
	UserStatusActive  = "active"
)

type User struct {
	ID        uuid.UUID
	Email     string
	Password  string
	Name      string
	Surename  string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}
