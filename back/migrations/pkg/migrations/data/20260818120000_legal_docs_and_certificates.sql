-- +goose Up
-- +goose StatementBegin

-- Юридические документы теперь версионируются файлами (PDF/скан), а не
-- только текстом: у документа есть текущий файл, у каждой версии — свой файл
-- (история файлов не удаляется при замене, только при удалении документа
-- целиком). Позволяет админке добавлять/удалять документы из перечня.
ALTER TABLE portal_legal_docs
    ADD COLUMN current_file_id UUID REFERENCES files (id);

ALTER TABLE portal_legal_doc_versions
    ADD COLUMN file_id UUID REFERENCES files (id);

-- Сертификаты продукции/компании с прикреплённым файлом (скан/фото).
CREATE TABLE portal_certificates (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title      VARCHAR(255) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    file_id    UUID REFERENCES files (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_portal_certificates_sort ON portal_certificates (is_active, sort_order);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_portal_certificates_sort;
DROP TABLE IF EXISTS portal_certificates;
ALTER TABLE portal_legal_doc_versions DROP COLUMN IF EXISTS file_id;
ALTER TABLE portal_legal_docs DROP COLUMN IF EXISTS current_file_id;
-- +goose StatementEnd
