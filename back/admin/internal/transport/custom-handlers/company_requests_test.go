package custom_handlers

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/mbatimel/AMC/admin/pkg/models"
)

type companyRequestRecorder struct {
	input models.CompanyRequestInput
	calls int
}

func (r *companyRequestRecorder) SendCompanyRequest(_ context.Context, input models.CompanyRequestInput) (models.CompanyRequestResponse, error) {
	r.calls++
	r.input = input
	return models.CompanyRequestResponse{Accepted: true}, nil
}

func TestCompanyRequestRouteParsesMultipart(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for field, value := range map[string]string{
		"contact_name": "Иван Иванов", "email": "ivan@example.com", "phone": "+79990000000",
		"company": "ООО Ромашка", "message": "Здравствуйте",
	} {
		if err := writer.WriteField(field, value); err != nil {
			t.Fatal(err)
		}
	}
	part, err := writer.CreateFormFile("attachments", "document.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = part.Write([]byte("%PDF-1.4\n")); err != nil {
		t.Fatal(err)
	}
	if err = writer.Close(); err != nil {
		t.Fatal(err)
	}

	recorder := &companyRequestRecorder{}
	app := fiber.New()
	NewCompanyRequestRoutes(zerolog.Nop(), recorder, 1024, 5).SetRoutes(app)
	request := httptest.NewRequest("POST", "/api/v1/company-requests", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 200 || recorder.calls != 1 {
		t.Fatalf("status=%d calls=%d", response.StatusCode, recorder.calls)
	}
	if recorder.input.Email != "ivan@example.com" || recorder.input.Message != "Здравствуйте" || len(recorder.input.Attachments) != 1 {
		t.Fatalf("input=%+v", recorder.input)
	}
}
