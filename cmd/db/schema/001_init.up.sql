CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email varchar(32) UNIQUE NOT NULL,
    password_hash varchar(60) NOT NULL
);
