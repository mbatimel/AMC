package clients

import (
	"context"

	"github.com/google/uuid"
	accessTransport "github.com/mbatimel/AMC/access/pkg/client/transport"
)

type AccessClient interface {
	CheckAccess(ctx context.Context, userID uuid.UUID, role int) (bool, error)
}

func NewAccessClient(address string) AccessClient {
	return accessTransport.NewClientAccessAPI(address)
}
