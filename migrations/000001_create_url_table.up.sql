CREATE TABLE IF NOT EXISTS urls (
    short_code   CHAR(10)     PRIMARY KEY,
    original_url TEXT         NOT NULL UNIQUE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);