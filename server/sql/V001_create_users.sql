CREATE TABLE IF NOT EXISTS users (
  id         BIGSERIAL PRIMARY KEY,
  name       TEXT NOT NULL,
  age        INTEGER NOT NULL CHECK (age >= 0),
  department TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_users_name ON users (name);
