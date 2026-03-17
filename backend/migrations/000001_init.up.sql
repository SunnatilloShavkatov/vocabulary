CREATE TABLE IF NOT EXISTS admins (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'admin',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS vocabularies (
    id UUID PRIMARY KEY,
    word TEXT NOT NULL,
    translation TEXT NOT NULL,
    example TEXT,
    created_by UUID NOT NULL REFERENCES admins(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_vocabularies_word ON vocabularies(word);
CREATE INDEX IF NOT EXISTS idx_vocabularies_translation ON vocabularies(translation);

