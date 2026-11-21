CREATE TABLE IF NOT EXISTS products (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255)   NOT NULL,
    description TEXT           NOT NULL DEFAULT '',
    price       NUMERIC(14, 2) NOT NULL CHECK (price >= 0),
    stock       INTEGER        NOT NULL DEFAULT 0 CHECK (stock >= 0),
    version     INTEGER        NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS flash_sale_slots (
    id          SERIAL PRIMARY KEY,
    product_id  INTEGER        NOT NULL REFERENCES products(id),
    flash_price NUMERIC(14, 2) NOT NULL CHECK (flash_price >= 0),
    quota       INTEGER        NOT NULL CHECK (quota >= 0),
    start_time  TIMESTAMPTZ    NOT NULL,
    end_time    TIMESTAMPTZ    NOT NULL,
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    CHECK (end_time > start_time)
);

CREATE INDEX IF NOT EXISTS idx_flash_sale_slots_product ON flash_sale_slots(product_id);
CREATE INDEX IF NOT EXISTS idx_flash_sale_slots_time ON flash_sale_slots(start_time, end_time);
