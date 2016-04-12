CREATE TABLE users (
    email TEXT PRIMARY KEY,
    token TEXT UNIQUE,
    registered boolean DEFAULT 0,
    requested DATETIME
);

CREATE TABLE namespaces (
    ns TEXT PRIMARY KEY,
    email TEXT REFERENCES users(email) ON DELETE CASCADE
);

CREATE TABLE packages (
    vcs TEXT,
    repo TEXT,
    path TEXT UNIQUE,
    ns TEXT REFERENCES namespaces(ns) ON DELETE CASCADE
);
