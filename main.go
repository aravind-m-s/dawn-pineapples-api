package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aravind-m-s/dawn-pineapples-api/db"
	"github.com/aravind-m-s/dawn-pineapples-api/handlers"
	"github.com/gorilla/mux"
)

func main() {
	db.InitDB()
	log.Print("starting server...")

	router := mux.NewRouter()

	// Define routes here
	// router.HandleFunc("/companies", getCompanies).Methods("GET")

	// Company routes
	router.HandleFunc("/companies", handlers.CreateCompany).Methods("POST")
	router.HandleFunc("/companies", handlers.GetCompanies).Methods("GET")
	router.HandleFunc("/companies/{id}", handlers.GetCompany).Methods("GET")
	router.HandleFunc("/companies/{id}", handlers.UpdateCompany).Methods("PUT")
	router.HandleFunc("/companies/{id}", handlers.DeleteCompany).Methods("DELETE")

	// Transaction routes
	router.HandleFunc("/transactions", handlers.CreateTransaction).Methods("POST")
	router.HandleFunc("/transactions", handlers.GetTransactions).Methods("GET")
	router.HandleFunc("/transactions/{id}", handlers.GetTransaction).Methods("GET")
	router.HandleFunc("/transactions/{id}", handlers.UpdateTransaction).Methods("PUT")
	router.HandleFunc("/transactions/{id}", handlers.DeleteTransaction).Methods("DELETE")

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("server listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}
