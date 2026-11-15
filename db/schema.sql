-- Jalankan sekali di Supabase (SQL Editor). Aman diulang.
-- Kalau teman satu tim sudah membuat tabel products, cocokkan kolomnya dengan ini.
CREATE TABLE IF NOT EXISTS products (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255)   NOT NULL,
    description TEXT           NOT NULL DEFAULT '',
    price       NUMERIC(14, 2) NOT NULL CHECK (price >= 0),
    stock       INTEGER        NOT NULL DEFAULT 0 CHECK (stock >= 0),
    version     INTEGER        NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);
