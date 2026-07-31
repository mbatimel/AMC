-- +goose Up
-- +goose StatementBegin

-- Контент публичных страниц портала (главная, о компании, условия, акции,
-- сертификаты, контакты). Редактируется в /admin/content/:pageKey.
CREATE TABLE portal_content_pages (
    page_key   VARCHAR(64) PRIMARY KEY,
    payload    JSONB NOT NULL,
    updated_by UUID,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Баннеры главной страницы. Порядок задаётся sort_order, период показа —
-- date_from / date_to (NULL = без ограничения).
CREATE TABLE portal_banners (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title      VARCHAR(255) NOT NULL,
    subtitle   TEXT NOT NULL DEFAULT '',
    image_url  TEXT NOT NULL DEFAULT '',
    link       TEXT NOT NULL DEFAULT '',
    sort_order INT NOT NULL DEFAULT 0,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    date_from  DATE,
    date_to    DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_portal_banners_sort ON portal_banners (is_active, sort_order);

-- Юридические документы с историей версий (удаление версий запрещено).
CREATE TABLE portal_legal_docs (
    id              VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,
    body            TEXT NOT NULL DEFAULT '',
    current_version VARCHAR(32) NOT NULL DEFAULT '1.0',
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE portal_legal_doc_versions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doc_id     VARCHAR(64) NOT NULL REFERENCES portal_legal_docs (id) ON DELETE CASCADE,
    version    VARCHAR(32) NOT NULL,
    summary    TEXT NOT NULL DEFAULT '',
    author     VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_portal_legal_versions_doc ON portal_legal_doc_versions (doc_id, created_at DESC);

-- Заявки на регистрацию (M-05): создаются при регистрации, модерируются
-- в /admin/signup-requests.
CREATE TABLE portal_signup_requests (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company       VARCHAR(255) NOT NULL DEFAULT '',
    inn           VARCHAR(32) NOT NULL DEFAULT '',
    contact       VARCHAR(255) NOT NULL DEFAULT '',
    email         VARCHAR(255) NOT NULL,
    phone         VARCHAR(64) NOT NULL DEFAULT '',
    request_type  VARCHAR(32) NOT NULL DEFAULT 'organization',
    status        VARCHAR(32) NOT NULL DEFAULT 'pending',
    reject_reason TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    decided_at    TIMESTAMPTZ
);

CREATE INDEX idx_portal_signup_requests_status ON portal_signup_requests (status, created_at DESC);

-- Обращения в поддержку: форма /support и вызов оператора из чата подбора.
CREATE TABLE portal_support_requests (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID,
    client_name VARCHAR(255) NOT NULL DEFAULT '',
    contact     VARCHAR(255) NOT NULL DEFAULT '',
    order_id    UUID,
    subject     VARCHAR(255) NOT NULL,
    body        TEXT NOT NULL,
    severity    SMALLINT NOT NULL DEFAULT 3,
    source      VARCHAR(32) NOT NULL DEFAULT 'form',
    status      VARCHAR(32) NOT NULL DEFAULT 'new',
    answer      TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_portal_support_requests_status ON portal_support_requests (status, created_at DESC);

-- Отзывы клиента по завершённым заказам.
CREATE TABLE portal_order_feedback (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id     UUID NOT NULL,
    user_id      UUID,
    client_name  VARCHAR(255) NOT NULL DEFAULT '',
    order_status VARCHAR(32) NOT NULL DEFAULT '',
    rating       SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    body         TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_portal_order_feedback_order ON portal_order_feedback (order_id);

-- Диалоги ИИ-помощника по подбору инструмента (M-04).
CREATE TABLE portal_assistant_messages (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL,
    user_id    UUID,
    author     VARCHAR(16) NOT NULL,
    body       TEXT NOT NULL,
    products   JSONB NOT NULL DEFAULT '[]'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_portal_assistant_messages_session
    ON portal_assistant_messages (session_id, created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_portal_assistant_messages_session;
DROP TABLE IF EXISTS portal_assistant_messages;
DROP INDEX IF EXISTS idx_portal_order_feedback_order;
DROP TABLE IF EXISTS portal_order_feedback;
DROP INDEX IF EXISTS idx_portal_support_requests_status;
DROP TABLE IF EXISTS portal_support_requests;
DROP INDEX IF EXISTS idx_portal_signup_requests_status;
DROP TABLE IF EXISTS portal_signup_requests;
DROP INDEX IF EXISTS idx_portal_legal_versions_doc;
DROP TABLE IF EXISTS portal_legal_doc_versions;
DROP TABLE IF EXISTS portal_legal_docs;
DROP INDEX IF EXISTS idx_portal_banners_sort;
DROP TABLE IF EXISTS portal_banners;
DROP TABLE IF EXISTS portal_content_pages;
-- +goose StatementEnd
