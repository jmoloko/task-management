-- Обновление схемы хранения паролей
ALTER TABLE users ALTER COLUMN password_hash TYPE VARCHAR(255);

-- Комментарий к колонке
COMMENT ON COLUMN users.password_hash IS 'Bcrypt hash of the password'; 