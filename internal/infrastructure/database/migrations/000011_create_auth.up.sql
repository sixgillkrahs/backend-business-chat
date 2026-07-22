CREATE TABLE IF NOT EXISTS auths (
    id UUID PRIMARY KEY,
    user_id INT NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role_id INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(20) DEFAULT 'system',
    updated_by VARCHAR(20) DEFAULT 'system',
    deleted_by VARCHAR(20) DEFAULT 'system',
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    is_deleted BOOLEAN DEFAULT false,
    CONSTRAINT fk_auth_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_auth_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);