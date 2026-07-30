package clients

import (
	"context"

	"github.com/google/uuid"

	accessTransport "github.com/mbatimel/AMC/access/pkg/client/transport"
)

type AccessClient interface {
	CheckAccess(ctx context.Context, userID uuid.UUID, role int) (allowed bool, err error)
	AddRole(ctx context.Context, adminUserID uuid.UUID, userID uuid.UUID, role int) (success bool, err error)
	UpdateRole(ctx context.Context, adminUserID uuid.UUID, userID uuid.UUID, role int) (success bool, err error)
}

func NewAccessClient(address string) AccessClient {
	return accessTransport.NewClientAccessAPI(address)
}
