CREATE TABLE books (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    published_at DATE,
    created_at TIMESTAMP DEFAULT NOW()
); 