package models

import (
	"time"

	"github.com/google/uuid"
)

type RoleCode int

const (
	RoleCodeAdmin    RoleCode = 0
	RoleCodeSupport  RoleCode = 1
	RoleCodeSupplier RoleCode = 2
	RoleCodeBuyer    RoleCode = 3
)

type Role struct {
	ID          uuid.UUID
	Code        RoleCode
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UserRole struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	RoleID    uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}
