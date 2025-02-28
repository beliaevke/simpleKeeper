package database

import "database/sql"

// CloseDB закрывает соединение с базой данных
func CloseDB(db *sql.DB) error {
	return db.Close()
}
