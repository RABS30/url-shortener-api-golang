CREATE TABLE short_urls (
    id              BIGSERIAL PRIMARY KEY NOT NULL,
    user_id         BIGINT NOT NULL,
    short_code      VARCHAR(255) NOT NULL,
    original_url    TEXT NOT NULL,
    expired_at      TIMESTAMP NULL,
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW(),


    CONSTRAINT fk_user 
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);