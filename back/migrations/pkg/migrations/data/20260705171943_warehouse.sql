-- +goose Up
-- +goose StatementBegin
CREATE TABLE warehouses (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    one_c_guid UUID UNIQUE,
    name VARCHAR(255),
    city VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (id)
);

CREATE TABLE stock_balances (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id),
    warehouse_id UUID REFERENCES warehouses(id),
    quantity NUMERIC,
    reserved_quantity NUMERIC,
    synced_at TIMESTAMPTZ,
    PRIMARY KEY (id)
);
CREATE INDEX idx_stock_balances_product_id ON stock_balances(product_id);
CREATE INDEX idx_stock_balances_warehouse_id ON stock_balances(warehouse_id);

CREATE TABLE stock_snapshots (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id),
    warehouse_id UUID REFERENCES warehouses(id),
    quantity NUMERIC,
    captured_at TIMESTAMPTZ,
    PRIMARY KEY (id)
);
CREATE INDEX idx_stock_snapshots_product_id ON stock_snapshots(product_id);
CREATE INDEX idx_stock_snapshots_warehouse_id ON stock_snapshots(warehouse_id);

CREATE TABLE stock_thresholds (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id),
    category_id UUID REFERENCES categories(id),
    threshold_quantity NUMERIC,
    manager_id UUID REFERENCES users(id),
    PRIMARY KEY (id)
);
CREATE INDEX idx_stock_thresholds_product_id ON stock_thresholds(product_id);
CREATE INDEX idx_stock_thresholds_category_id ON stock_thresholds(category_id);
CREATE INDEX idx_stock_thresholds_manager_id ON stock_thresholds(manager_id);

CREATE TABLE replenishment_recommendations (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id),
    warehouse_id UUID REFERENCES warehouses(id),
    recommended_quantity NUMERIC,
    recommended_order_date DATE,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
CREATE INDEX idx_replenishment_recommendations_product_id ON replenishment_recommendations(product_id);
CREATE INDEX idx_replenishment_recommendations_warehouse_id ON replenishment_recommendations(warehouse_id);

CREATE TABLE sales_history (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id),
    category_id UUID REFERENCES categories(id),
    counterparty_id UUID REFERENCES counterparties(id),
    region VARCHAR(255),
    period_date DATE,
    quantity NUMERIC,
    amount NUMERIC,
    source VARCHAR(255),
    PRIMARY KEY (id)
);
CREATE INDEX idx_sales_history_product_id ON sales_history(product_id);
CREATE INDEX idx_sales_history_category_id ON sales_history(category_id);
CREATE INDEX idx_sales_history_counterparty_id ON sales_history(counterparty_id);

CREATE TABLE demand_forecasts (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id),
    forecast_date DATE,
    horizon_months INTEGER,
    predicted_quantity NUMERIC,
    confidence_low NUMERIC,
    confidence_high NUMERIC,
    mape NUMERIC,
    model_version VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
CREATE INDEX idx_demand_forecasts_product_id ON demand_forecasts(product_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS demand_forecasts CASCADE;
DROP TABLE IF EXISTS sales_history CASCADE;
DROP TABLE IF EXISTS replenishment_recommendations CASCADE;
DROP TABLE IF EXISTS stock_thresholds CASCADE;
DROP TABLE IF EXISTS stock_snapshots CASCADE;
DROP TABLE IF EXISTS stock_balances CASCADE;
DROP TABLE IF EXISTS warehouses CASCADE;
-- +goose StatementEnd
