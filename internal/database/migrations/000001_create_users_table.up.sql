-- Active: 1780514042443@@127.0.0.1@5432@url_shorter

CREATE TYPE user_status AS ENUM ('active', 'suspended', 'inactive');

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    is_verified BOOLEAN DEFAULT FALSE,
    status user_status DEFAULT 'active', 
    created_at TIMESTAMP DEFAULT NOW() 
);
