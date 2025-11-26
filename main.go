package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"programming_naming/db"
	"programming_naming/function"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	//test時には、DBを使用しない
	if os.Getenv("ENV") == "test" {
		log.Println("Test environment detected, skipping database connection")
		fmt.Println(os.Getenv("testcode"))
		return
	}

	// 1. データベース接続文字列を環境変数から取得
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set in .env file or environment variables")
	}

	// 2. データベース接続を初期化 (db.NewDBを呼び出す)
	dbConn, err := db.NewDB(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()
	log.Println("Successfully connected to database!")

	// 3. マイグレーションを実行
	if err := db.Migrate(dbConn); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Serviceの初期化
	funcService := function.NewService(dbConn)

	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/index.html")
	}).Methods("GET")

	// データベース接続(dbConn)をハンドラに渡す
	r.HandleFunc("/function/{id}", funcService.GetFunctionById).Methods("GET")

	r.HandleFunc("/functions", funcService.GetAllFunctions).Methods("GET")

	log.Println("Server starting at :8080")
	http.ListenAndServe(":8080", r)
}
