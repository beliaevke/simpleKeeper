-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS Users (
    userID int primary key generated always as identity,
    userLogin varchar(200) not null unique,
    userPassword varchar(200) not null,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS Secrets(
    id int primary key generated always as identity,
    name varchar(255) not null unique,
    type varchar(10) not null,
    content bytea not null,
    version UUID default uuid_generate_v4() not null unique,
    ownerID integer references Users (userID),
    keyID integer references Keys (keyID),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    UNIQUE (name, ownerID)
);

CREATE TABLE IF NOT EXISTS Keys (
    keyID int primary key,
    keyAES VARCHAR NOT NULL,
    ownerID BIGINT NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    UNIQUE (keyID, ownerID)
);

-- +goose Down
DROP TABLE IF EXISTS Users;
DROP TABLE IF EXISTS Secrets;
DROP TABLE IF EXISTS Keys;