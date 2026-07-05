CREATE TABLE user_otps (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    otp_code VARCHAR(6) NOT NULL,
    type VARCHAR(20) NOT NULL,
    is_used BOOLEAN DEFAULT FALSE,
    expired_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_user_otps_email_code_type ON user_otps(email, otp_code, type);