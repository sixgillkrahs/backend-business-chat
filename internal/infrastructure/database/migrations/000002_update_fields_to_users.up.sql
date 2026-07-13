ALTER TABLE users ADD COLUMN full_name VARCHAR(100);
ALTER TABLE users ADD COLUMN avatar_url VARCHAR(255);
ALTER TABLE users ADD COLUMN phone_number VARCHAR(20);
ALTER TABLE users ADD COLUMN address TEXT;
ALTER TABLE users ADD COLUMN created_by VARCHAR(20);
ALTER TABLE users ADD COLUMN updated_by VARCHAR(20);
ALTER TABLE users ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN deleted_by VARCHAR(20);
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN is_active BOOLEAN DEFAULT true;
ALTER TABLE users ADD COLUMN is_deleted BOOLEAN DEFAULT false;

-- Normalize existing data
UPDATE users SET full_name = username WHERE full_name IS NULL;
UPDATE users SET is_deleted = false WHERE is_deleted IS NULL;
UPDATE users SET updated_at = created_at WHERE updated_at IS NULL;
UPDATE users SET deleted_at = created_at WHERE deleted_at IS NULL;
UPDATE users SET created_by = 'admin' WHERE created_by IS NULL;
UPDATE users SET updated_by = 'admin' WHERE updated_by IS NULL;
UPDATE users SET deleted_by = 'admin' WHERE deleted_by IS NULL;

-- Set NOT NULL constraint after populating data
ALTER TABLE users ALTER COLUMN full_name SET NOT NULL;