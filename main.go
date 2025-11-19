package main

import (
	"fmt"
	"net/http"
	"programming_naming/function"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/index.html")
	}).Methods("GET")

	r.HandleFunc("/function/{id}", function.GetFunctionById).Methods("GET")

	r.HandleFunc("/functions", function.GetAllFunctions).Methods("GET")

	fmt.Println("Server starting at :8080")
	http.ListenAndServe(":8080", r)
}
