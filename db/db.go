package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQLドライバーをインポート
)

// NewDB はデータベース接続を初期化し、*sql.DB インスタンスを返します。
// dataSourceName は "user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local" のような形式です。
func NewDB(dataSourceName string) (*sql.DB, error) {
	// "mysql" ドライバを使用してデータベースに接続
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// データベースへの接続を検証します
	if err = db.Ping(); err != nil {
		db.Close() // Pingに失敗したら接続を閉じます
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// パフォーマンス向上のためのコネクションプールの設定
	db.SetMaxOpenConns(25)                 // 同時に開かれる接続の最大数
	db.SetMaxIdleConns(25)                 // アイドル状態の接続の最大数
	db.SetConnMaxLifetime(5 * time.Minute) // 接続が再利用される最大時間

	return db, nil
}
