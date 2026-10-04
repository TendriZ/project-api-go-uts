CREATE TABLE IF NOT EXISTS roles (
    name VARCHAR(20) PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS permissions (
    name VARCHAR(50) PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_name       VARCHAR(20) NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
    permission_name VARCHAR(50) NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

-- Seed roles
INSERT INTO roles (name) VALUES ('admin'), ('staff'), ('user')
ON CONFLICT DO NOTHING;

-- Seed permissions
INSERT INTO permissions (name) VALUES
    ('book:list'),
    ('book:read:any'),
    ('book:create'),
    ('book:update:any'),
    ('book:delete'),
    ('loan:list:any'),
    ('loan:read:any'),
    ('loan:create'),
    ('loan:return'),
    ('loan:delete'),
    ('role:assign'),
    ('user:list'),
    ('user:read:any'),
    ('user:create'),
    ('user:update:any'),
    ('user:delete')
ON CONFLICT DO NOTHING;

-- admin: semua permission
INSERT INTO role_permissions (role_name, permission_name)
SELECT 'admin', name FROM permissions
ON CONFLICT DO NOTHING;

-- staff permissions
INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('staff', 'book:list'),
    ('staff', 'book:read:any'),
    ('staff', 'book:create'),
    ('staff', 'book:update:any'),
    ('staff', 'loan:list:any'),
    ('staff', 'loan:read:any'),
    ('staff', 'loan:create'),
    ('staff', 'loan:return'),
    ('staff', 'user:list'),
    ('staff', 'user:read:any')
ON CONFLICT DO NOTHING;

-- user permissions
INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('user', 'book:list'),
    ('user', 'book:read:any')
ON CONFLICT DO NOTHING;

-- FK constraint: users.role → roles.name
ALTER TABLE users ADD CONSTRAINT users_role_fk
    FOREIGN KEY (role) REFERENCES roles(name);
