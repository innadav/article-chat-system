CREATE TABLE articles (
    url TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    excerpt TEXT,
    summary TEXT,
    sentiment TEXT,
    topics TEXT[],
    entities TEXT[],
    processed_at TIMESTAMPTZ NOT NULL
);
