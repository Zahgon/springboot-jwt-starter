// Package db owns the embedded database: schema creation and seed loading,
// which Hibernate's ddl-auto and Spring's import.sql handling did before.
package db

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "modernc.org/sqlite" // pure-Go driver, no cgo
)

//go:embed import.sql
var importSQL string

// schema mirrors the JPA entity mappings: @Table(name="USERS"),
// @Table(name="AUTHORITY") and the @JoinTable named "user_authority".
const schema = `
CREATE TABLE USERS (
    id                       INTEGER PRIMARY KEY AUTOINCREMENT,
    username                 TEXT,
    password                 TEXT,
    first_name               TEXT,
    last_name                TEXT,
    email                    TEXT,
    phone_number             TEXT,
    enabled                  BOOLEAN,
    last_password_reset_date TEXT
);
CREATE TABLE AUTHORITY (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT
);
CREATE TABLE USER_AUTHORITY (
    user_id      INTEGER NOT NULL,
    authority_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, authority_id)
);
`

// Open creates a fresh in-memory database, applies the schema and loads the
// seed rows. Each call yields an independent database, so tests do not share
// state. The shared-cache URI plus a single connection keeps every statement
// on the same in-memory instance.
func Open(name string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", name)
	database, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("db: opening %s: %w", dsn, err)
	}
	// An in-memory database lives only as long as a connection to it is open.
	database.SetMaxOpenConns(1)

	if _, err := database.Exec(schema); err != nil {
		database.Close()
		return nil, fmt.Errorf("db: creating schema: %w", err)
	}
	if _, err := database.Exec(importSQL); err != nil {
		database.Close()
		return nil, fmt.Errorf("db: loading seed data: %w", err)
	}
	return database, nil
}
