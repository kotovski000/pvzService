CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('employee', 'moderator')),
    created_at TIMESTAMP DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS pvz (
    id UUID PRIMARY KEY,
    city TEXT NOT NULL CHECK (city IN ('Москва', 'Санкт-Петербург', 'Казань')),
    registration_date TIMESTAMP DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS receptions (
    id UUID PRIMARY KEY,
    pvz_id UUID REFERENCES pvz(id),
    status TEXT NOT NULL CHECK (status IN ('in_progress', 'close')),
    created_at TIMESTAMP DEFAULT NOW(),
    closed_at TIMESTAMP
    );

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY,
    reception_id UUID REFERENCES receptions(id),
    type TEXT NOT NULL CHECK (type IN ('электроника', 'одежда', 'обувь')),
    created_at TIMESTAMP DEFAULT NOW()
    );

