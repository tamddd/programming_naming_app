package db

import (
	"database/sql"
	"fmt"
)

// Migrate はデータベースのマイグレーション（テーブル作成など）を実行します。
func Migrate(db *sql.DB) error {
	if err := createUserTable(db); err != nil {
		return err
	}
	if err := createFunctionTable(db); err != nil {
		return err
	}
	fmt.Println("Database migration completed successfully!")
	return nil
}

func createUserTable(db *sql.DB) error {
	// IF NOT EXISTS を使うことで、テーブルが既に存在していてもエラーになりません。
	query := `
    CREATE TABLE IF NOT EXISTS users (
        id BIGINT PRIMARY KEY AUTO_INCREMENT,
        name VARCHAR(255) NOT NULL,
        email VARCHAR(255) NOT NULL UNIQUE,
        password VARCHAR(255) NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    );`

	_, err := db.Exec(query)
	return err
}

func createFunctionTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS functions (
        id BIGINT PRIMARY KEY AUTO_INCREMENT,
        title VARCHAR(255) NOT NULL,
        description TEXT,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        is_public BOOLEAN NOT NULL DEFAULT FALSE,
        user_id BIGINT NOT NULL,
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    );`

	_, err := db.Exec(query)
	return err
}
