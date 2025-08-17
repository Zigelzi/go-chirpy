-- Project setup documentation
CREATE DATABASE chirpy;

CREATE USER chirpy_app_user
WITH
    PASSWORD 'dev_password';

GRANT CONNECT ON DATABASE chirpy TO chirpy_app_user;

GRANT USAGE,
CREATE ON SCHEMA public TO chirpy_app_user;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO chirpy_app_user;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL PRIVILEGES ON TABLES TO chirpy_app_user;