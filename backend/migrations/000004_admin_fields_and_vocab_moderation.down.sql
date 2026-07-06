DROP INDEX IF EXISTS idx_vocabularies_category;
DROP INDEX IF EXISTS idx_vocabularies_status;
DROP INDEX IF EXISTS idx_users_status;

ALTER TABLE vocabularies
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS category;

ALTER TABLE users
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS name;
