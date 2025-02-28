CREATE TABLE IF NOT EXISTS Secrets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL,
    content BLOB NOT NULL,
    version TEXT DEFAULT (lower(hex(randomblob(16)))) NOT NULL UNIQUE,
    ownerID INTEGER,
    keyID INTEGER,
    FOREIGN KEY (ownerID) REFERENCES Users (userID),
    FOREIGN KEY (keyID) REFERENCES Keys (keyID),
    UNIQUE (name, ownerID)
);