package clients

import (
	"context"

	"github.com/google/uuid"
	productsTransport "github.com/mbatimel/AMC/products/pkg/client/transport"
)

// ProductsClient — вызовы внутреннего (service-to-service) API products.
type ProductsClient interface {
	SetMarketplaceStatus(ctx context.Context, productID uuid.UUID, marketplace string, status string) (bool, error)
}

func NewProductsClient(address string) ProductsClient {
	return productsTransport.NewClientProductsInternalAPI(address)
}
