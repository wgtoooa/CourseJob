CREATE TABLE auth_refresh_token (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_id TEXT NOT NULL UNIQUE,
    token_hash TEXT NOT NULL UNIQUE,
    issued_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    replaced_by_token_id TEXT,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (expires_at > issued_at)
);

CREATE INDEX idx_auth_refresh_token_user_id ON auth_refresh_token(user_id);
CREATE INDEX idx_auth_refresh_token_expires_at ON auth_refresh_token(expires_at);
CREATE INDEX idx_auth_refresh_token_revoked_at ON auth_refresh_token(revoked_at);
