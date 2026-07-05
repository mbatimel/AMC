-- +goose Up
-- +goose StatementBegin
CREATE TABLE ai_chat_sessions (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    counterparty_id UUID REFERENCES counterparties(id),
    status VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    closed_at TIMESTAMPTZ,
    PRIMARY KEY (id)
);
CREATE INDEX idx_ai_chat_sessions_user_id ON ai_chat_sessions(user_id);
CREATE INDEX idx_ai_chat_sessions_counterparty_id ON ai_chat_sessions(counterparty_id);

CREATE TABLE ai_chat_messages (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    session_id UUID REFERENCES ai_chat_sessions(id),
    role VARCHAR(255),
    content TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
CREATE INDEX idx_ai_chat_messages_session_id ON ai_chat_messages(session_id);

CREATE TABLE ai_product_recommendations (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    session_id UUID REFERENCES ai_chat_sessions(id),
    product_id UUID REFERENCES products(id),
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
CREATE INDEX idx_ai_product_recommendations_session_id ON ai_product_recommendations(session_id);
CREATE INDEX idx_ai_product_recommendations_product_id ON ai_product_recommendations(product_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS ai_product_recommendations CASCADE;
DROP TABLE IF EXISTS ai_chat_messages CASCADE;
DROP TABLE IF EXISTS ai_chat_sessions CASCADE;
-- +goose StatementEnd
