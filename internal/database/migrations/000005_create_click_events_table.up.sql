CREATE TABLE click_events (
    id            BIGSERIAL PRIMARY KEY,
    short_url_id  BIGINT NOT NULL,
    ip_address    VARCHAR(255),
    user_agent    TEXT,
    referer       TEXT,
    clicked_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_short_url
        FOREIGN KEY (short_url_id)
        REFERENCES short_urls(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_click_events_short_url_id ON click_events(short_url_id);