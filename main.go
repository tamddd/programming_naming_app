package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"programming_naming/db"
	"programming_naming/function"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func basicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()

		if !ok || username != "admin" || password != "password" {
			w.Header().Set("WWW-Authenticate", `Basic realm="Admin Area"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	var dbConn *sql.DB
	//test時には、DBを使用しない
	if os.Getenv("ENV") == "test" {
		dbConn = nil
		log.Println("Test environment detected, skipping database connection")
		return
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set in .env file or environment variables")
	}

	dbConn, err := db.NewDB(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()
	log.Println("Successfully connected to database!")

	if err := db.Migrate(dbConn); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	funcService := function.NewService(dbConn)

	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/index.html")
	}).Methods("GET")

	r.HandleFunc("/function/{id}", funcService.GetFunctionById).Methods("GET")
	r.HandleFunc("/functions", funcService.GetAllFunctions).Methods("GET")

	// 管理者用: 問題作成（Basic認証が必要）
	adminRouter := r.PathPrefix("/admin").Subrouter()
	adminRouter.Use(func(next http.Handler) http.Handler {
		return basicAuthMiddleware(next)
	})
	adminRouter.HandleFunc("/functions/new", funcService.NewFunctionForm).Methods("GET")
	adminRouter.HandleFunc("/functions/generate", funcService.GenerateWithAI).Methods("POST")

	// POST /functions は認証不要（フォーム送信先）
	r.HandleFunc("/functions", funcService.CreateFunction).Methods("POST")

	log.Println("Server starting at :8080")
	http.ListenAndServe(":8080", r)
}
