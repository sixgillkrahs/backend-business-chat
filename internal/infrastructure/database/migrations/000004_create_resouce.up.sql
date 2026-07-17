CREATE TABLE IF NOT EXISTS resources (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO resources (name, code, description) VALUES
('Quản lý người dùng', 'USER_MANAGEMENT', 'Chức năng xem, tạo, sửa, xóa người dùng'),
('Quản lý phân quyền', 'ROLE_MANAGEMENT', 'Chức năng cấu hình vai trò và quyền hạn'),
('Cấu hình hệ thống', 'SYSTEM_SETTINGS', 'Chức năng điều chỉnh các thông số hệ thống')
ON CONFLICT (code) DO NOTHING