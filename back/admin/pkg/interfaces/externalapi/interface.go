// Package externalapi describes the public admin panel API contract.
// @tg version=0.0.1
// @tg backend=admin
// @tg title=`admin`
// @tg servers=
//
//go:generate tg transport --services . --out ../../../internal/transport/jsonRPC/externalapi --outSwagger ../../../swaggers/externalapi/swagger.yaml
package externalapi

import (
	"context"

	"github.com/google/uuid"
	"github.com/mbatimel/AMC/admin/pkg/models"
)

// AdminAPI
// @tg http-server metrics log
// @tg http-prefix=/api
// @tg 200=github.com/mbatimel/AMC/admin/swaggers/externalapi/models:Resp200
// @tg 400=github.com/mbatimel/AMC/admin/swaggers/externalapi/models:Err400
// @tg 401=github.com/mbatimel/AMC/admin/swaggers/externalapi/models:Err401
// @tg 403=github.com/mbatimel/AMC/admin/swaggers/externalapi/models:Err403
// @tg 500=github.com/mbatimel/AMC/admin/swaggers/externalapi/models:Err500
type AdminAPI interface {
	// Login ...
	// @tg http-method=POST
	// @tg http-path=/v1/admin/auth/login
	// @tg http-args=email|email
	// @tg http-args=password|password
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:Login
	// @tg summary=`Вход в админ-панель`
	// @tg desc=`Авторизация сотрудника портала по email и паролю`
	// @tg uuidPackage=github.com/google/uuid
	Login(ctx context.Context, email string, password string) (response models.LoginResponse, err error)

	// Logout ...
	// @tg http-method=POST
	// @tg http-path=/v1/admin/auth/logout
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:Logout
	// @tg summary=`Выход из админ-панели`
	// @tg desc=`Завершение сессии сотрудника портала`
	// @tg uuidPackage=github.com/google/uuid
	Logout(ctx context.Context, userID uuid.UUID) (err error)

	// GetSession ...
	// @tg http-method=GET
	// @tg http-path=/v1/admin/auth/session
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:GetSession
	// @tg summary=`Проверка сессии`
	// @tg desc=`Проверка, что пользователь авторизован и имеет права администратора`
	// @tg uuidPackage=github.com/google/uuid
	GetSession(ctx context.Context, userID uuid.UUID) (response models.SessionResponse, err error)

	// ListAuditLog ...
	// @tg http-method=GET
	// @tg http-path=/v1/admin/audit-log
	// @tg http-headers=userID|X-User-Id
	// @tg http-args=limit|limit
	// @tg http-args=offset|offset
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:ListAuditLog
	// @tg summary=`Журнал действий`
	// @tg desc=`Список действий администраторов портала без возможности удаления`
	// @tg uuidPackage=github.com/google/uuid
	ListAuditLog(ctx context.Context, userID uuid.UUID, limit int, offset int) (response models.ListAuditLogResponse, err error)

	// CreateSignupRequest ...
	// @tg http-method=POST
	// @tg http-path=/v1/signup-requests
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:CreateSignupRequest
	// @tg summary=`Заявка на регистрацию`
	// @tg desc=`Публичное создание заявки на регистрацию; модерируется в /v1/admin/signup-requests`
	CreateSignupRequest(ctx context.Context, company string, inn string, contact string, email string, phone string, requestType string) (response models.SignupRequest, err error)

	// ListSignupRequests ...
	// @tg http-method=GET
	// @tg http-path=/v1/admin/signup-requests
	// @tg http-headers=userID|X-User-Id
	// @tg http-args=status|status
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:ListSignupRequests
	// @tg summary=`Список заявок на регистрацию`
	// @tg desc=`Возвращает заявки на регистрацию, опционально отфильтрованные по статусу`
	// @tg uuidPackage=github.com/google/uuid
	ListSignupRequests(ctx context.Context, userID uuid.UUID, status string) (response models.ListSignupRequestsResponse, err error)

	// ApproveSignupRequest ...
	// @tg http-method=POST
	// @tg http-path=/v1/admin/signup-requests/:requestID/approve
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:ApproveSignupRequest
	// @tg summary=`Одобрение заявки`
	// @tg desc=`Помечает заявку одобренной; аккаунт уже существует (регистрация мгновенная), ничего не создаётся`
	// @tg uuidPackage=github.com/google/uuid
	// @tg requestID.format=uuid
	ApproveSignupRequest(ctx context.Context, userID uuid.UUID, requestID uuid.UUID) (response models.SignupRequest, err error)

	// RejectSignupRequest ...
	// @tg http-method=POST
	// @tg http-path=/v1/admin/signup-requests/:requestID/reject
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:RejectSignupRequest
	// @tg summary=`Отклонение заявки`
	// @tg desc=`Отклоняет заявку, удаляет аккаунт заявителя в users (если найден по email) и отправляет письмо с причиной отказа`
	// @tg uuidPackage=github.com/google/uuid
	// @tg requestID.format=uuid
	RejectSignupRequest(ctx context.Context, userID uuid.UUID, requestID uuid.UUID, reason string) (response models.SignupRequest, err error)

	// ListLegalDocs ...
	// @tg http-method=GET
	// @tg http-path=/v1/admin/legal-docs
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:ListLegalDocs
	// @tg summary=`Список документов и соглашений`
	// @tg desc=`Возвращает перечень юридических документов (админка)`
	// @tg uuidPackage=github.com/google/uuid
	ListLegalDocs(ctx context.Context, userID uuid.UUID) (response models.ListLegalDocsResponse, err error)

	// ListPublicLegalDocs ...
	// @tg http-method=GET
	// @tg http-path=/v1/legal-docs
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:ListPublicLegalDocs
	// @tg summary=`Публичный список документов и соглашений`
	// @tg desc=`Возвращает перечень юридических документов для витрины портала`
	ListPublicLegalDocs(ctx context.Context) (response models.ListLegalDocsResponse, err error)

	// CreateLegalDoc ...
	// @tg http-method=POST
	// @tg http-path=/v1/admin/legal-docs
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:CreateLegalDoc
	// @tg summary=`Добавление документа`
	// @tg desc=`Добавляет новый документ в перечень «Документы и соглашения» вместе с первой версией файла (fileContentBase64 — файл, закодированный в base64)`
	// @tg uuidPackage=github.com/google/uuid
	CreateLegalDoc(ctx context.Context, userID uuid.UUID, id string, name string, version string, summary string, fileName string, fileContentBase64 string) (response models.LegalDoc, err error)

	// ReplaceLegalDocFile ...
	// @tg http-method=PATCH
	// @tg http-path=/v1/admin/legal-docs/:docID
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:ReplaceLegalDocFile
	// @tg summary=`Замена файла документа`
	// @tg desc=`Загружает новую версию файла документа («Заменить»); прежние версии сохраняются в истории (fileContentBase64 — файл, закодированный в base64)`
	// @tg uuidPackage=github.com/google/uuid
	ReplaceLegalDocFile(ctx context.Context, userID uuid.UUID, docID string, version string, summary string, fileName string, fileContentBase64 string) (response models.LegalDoc, err error)

	// DeleteLegalDoc ...
	// @tg http-method=DELETE
	// @tg http-path=/v1/admin/legal-docs/:docID
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:DeleteLegalDoc
	// @tg summary=`Удаление документа`
	// @tg desc=`Удаляет документ из перечня «Документы и соглашения» вместе со всей историей версий и файлами`
	// @tg uuidPackage=github.com/google/uuid
	DeleteLegalDoc(ctx context.Context, userID uuid.UUID, docID string) (response models.DeleteLegalDocResponse, err error)

	// ListLegalDocVersions ...
	// @tg http-method=GET
	// @tg http-path=/v1/admin/legal-docs/:docID/versions
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:ListLegalDocVersions
	// @tg summary=`История версий документа`
	// @tg desc=`Возвращает историю версий документа («История»)`
	// @tg uuidPackage=github.com/google/uuid
	ListLegalDocVersions(ctx context.Context, userID uuid.UUID, docID string) (response models.ListLegalDocVersionsResponse, err error)

	// ListCertificates ...
	// @tg http-method=GET
	// @tg http-path=/v1/admin/certificates
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:ListCertificates
	// @tg summary=`Список сертификатов`
	// @tg desc=`Возвращает перечень сертификатов (админка)`
	// @tg uuidPackage=github.com/google/uuid
	ListCertificates(ctx context.Context, userID uuid.UUID) (response models.ListCertificatesResponse, err error)

	// ListPublicCertificates ...
	// @tg http-method=GET
	// @tg http-path=/v1/certificates
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:ListPublicCertificates
	// @tg summary=`Публичный список сертификатов`
	// @tg desc=`Возвращает активные сертификаты для витрины портала`
	ListPublicCertificates(ctx context.Context) (response models.ListCertificatesResponse, err error)

	// CreateCertificate ...
	// @tg http-method=POST
	// @tg http-path=/v1/admin/certificates
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:CreateCertificate
	// @tg summary=`Добавление сертификата`
	// @tg desc=`Добавляет сертификат с прикреплённым файлом (fileContentBase64 — файл, закодированный в base64)`
	// @tg uuidPackage=github.com/google/uuid
	CreateCertificate(ctx context.Context, userID uuid.UUID, title string, sortOrder int, isActive bool, fileName string, fileContentBase64 string) (response models.Certificate, err error)

	// UpdateCertificate ...
	// @tg http-method=PATCH
	// @tg http-path=/v1/admin/certificates/:certID
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:UpdateCertificate
	// @tg summary=`Изменение сертификата`
	// @tg desc=`Обновляет сертификат; при непустых fileName/fileContentBase64 заменяет прикреплённый файл`
	// @tg uuidPackage=github.com/google/uuid
	// @tg certID.format=uuid
	UpdateCertificate(ctx context.Context, userID uuid.UUID, certID uuid.UUID, title string, sortOrder int, isActive bool, fileName string, fileContentBase64 string) (response models.Certificate, err error)

	// DeleteCertificate ...
	// @tg http-method=DELETE
	// @tg http-path=/v1/admin/certificates/:certID
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:DeleteCertificate
	// @tg summary=`Удаление сертификата`
	// @tg desc=`Удаляет сертификат и прикреплённый файл`
	// @tg uuidPackage=github.com/google/uuid
	// @tg certID.format=uuid
	DeleteCertificate(ctx context.Context, userID uuid.UUID, certID uuid.UUID) (response models.DeleteCertificateResponse, err error)
}
