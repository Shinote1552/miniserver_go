CREATE TABLE IF NOT EXISTS users(
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS urls (
    id BIGSERIAL PRIMARY KEY,
    short_key VARCHAR(10) UNIQUE NOT NULL,
    original_url TEXT NOT NULL UNIQUE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_urls_user_id ON urls(user_id) WHERE is_deleted = false;
CREATE INDEX IF NOT EXISTS idx_urls_short_key ON urls(short_key)WHERE is_deleted = false;
CREATE INDEX IF NOT EXISTS idx_urls_original_url ON urls(original_url)WHERE is_deleted = false;
CREATE INDEX IF NOT EXISTS idx_urls_is_deleted ON urls(is_deleted);
