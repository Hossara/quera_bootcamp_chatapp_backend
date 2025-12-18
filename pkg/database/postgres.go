package database

import (
	"database/sql"
	"fmt"

	"github.com/Hossara/quera_bootcamp_chatapp_backend/config"
	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

func openSQLDB(cfg config.PostgresDBConfig) (*sql.DB, error) {
	conn, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return conn, nil
}

func New(cfg config.PostgresDBConfig) (*DB, error) {
	conn, err := openSQLDB(cfg)
	if err != nil {
		return nil, err
	}

	return &DB{conn: conn}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) GetConnection() *sql.DB {
	return db.conn
}
