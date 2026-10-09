-- migrate:up

CREATE TABLE users (
  id UUID PRIMARY KEY,
  email VARCHAR(254) NOT NULL UNIQUE CHECK (email = lower(email)),
  password_hash TEXT NOT NULL,
  display_name VARCHAR(100) NOT NULL,
  email_verified_at TIMESTAMPTZ,
  created TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE sessions (
  token_hash BYTEA PRIMARY KEY CHECK (octet_length(token_hash) = 32),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at TIMESTAMPTZ NOT NULL,
  created TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX sessions_user_id_idx ON sessions(user_id);
CREATE INDEX sessions_expires_at_idx ON sessions(expires_at);

CREATE TABLE account_tokens (
  token_hash BYTEA PRIMARY KEY CHECK (octet_length(token_hash) = 32),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  purpose TEXT NOT NULL CHECK (purpose IN ('verify_email', 'reset_password')),
  expires_at TIMESTAMPTZ NOT NULL,
  created TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX account_tokens_user_id_purpose_idx ON account_tokens(user_id, purpose);
CREATE INDEX account_tokens_expires_at_idx ON account_tokens(expires_at);

-- Ownership is nullable during the explicit legacy-data backfill.
ALTER TABLE tracklists ADD COLUMN owner_id UUID REFERENCES users(id);
CREATE INDEX tracklists_owner_id_date_idx ON tracklists(owner_id, date DESC);

-- migrate:down

ALTER TABLE tracklists DROP COLUMN owner_id;
DROP TABLE account_tokens;
DROP TABLE sessions;
DROP TABLE users;
