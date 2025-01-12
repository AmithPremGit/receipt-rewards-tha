package main

import (
	"log"      // log
	"net/http" // net/http

	"github.com/gorilla/mux" // gorilla/mux
)

func main() {

	r := mux.NewRouter()

	// Initialize handlers
	h := NewHandlers()

	// Set up routes
	r.HandleFunc("/receipts/process", h.ProcessReceipt).Methods("POST")
	r.HandleFunc("/receipts/{id}/points", h.GetPoints).Methods("GET")

	// Start server on port 8080
	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))

}
