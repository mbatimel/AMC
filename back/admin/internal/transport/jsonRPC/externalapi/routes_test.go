package externalapi_test

import (
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	adminService "github.com/mbatimel/AMC/admin/internal/service"
	transport "github.com/mbatimel/AMC/admin/internal/transport/jsonRPC/externalapi"
)

func TestSignupRequestRoutesAreNotRegistered(t *testing.T) {
	svc := adminService.NewAdminApiService(zerolog.Nop(), nil, nil, nil, nil)
	server := transport.New(zerolog.Nop(), transport.AdminAPI(transport.NewAdminAPI(svc)))

	tests := []struct {
		method string
		path   string
	}{
		{method: "POST", path: "/api/v1/signup-requests"},
		{method: "GET", path: "/api/v1/admin/signup-requests"},
		{method: "POST", path: "/api/v1/admin/signup-requests/00000000-0000-0000-0000-000000000001/approve"},
		{method: "POST", path: "/api/v1/admin/signup-requests/00000000-0000-0000-0000-000000000001/reject"},
	}

	for _, test := range tests {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			response, err := server.Fiber().Test(httptest.NewRequest(test.method, test.path, nil))
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if response.StatusCode != 404 {
				t.Fatalf("status = %d, want 404", response.StatusCode)
			}
		})
	}
}
