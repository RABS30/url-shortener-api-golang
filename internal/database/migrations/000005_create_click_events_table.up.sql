CREATE TABLE click_events (
    id  BIGSERIAL PRIMARY KEY,
    short_url_id BIGINT NOT NULL,
    ip_address VARCHAR(255),
    user_agent TEXT,
    referer TEXT,
    clicked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);