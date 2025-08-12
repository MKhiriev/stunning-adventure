CREATE TABLE IF NOT EXISTS metrics (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    delta INT,
    value DOUBLE PRECISION
);