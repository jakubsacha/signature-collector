package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jakubsacha/signature-collector/handlers"
	"github.com/jakubsacha/signature-collector/i18n"
	"github.com/jakubsacha/signature-collector/models"
	"github.com/joho/godotenv"
)

func main() {
	// Set up logging
	log.Println("Starting application...")

	// Load .env file
	log.Println("Loading .env file...")
	err := godotenv.Load()
	if err != nil {
		log.Println("Didn't load .env file")
	}
	log.Println(".env file loaded successfully")

	// Initialize i18n
	log.Println("Initializing i18n...")
	err = i18n.Init(os.Getenv("LANGUAGE"))
	if err != nil {
		log.Fatalf("Error initializing i18n: %v", err)
	}
	log.Println("i18n initialized successfully")

	// Initialize the database
	log.Println("Setting up database configuration...")
	var config models.DBConfig
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		log.Println("Using MySQL database configuration")
		// Use MySQL if DB_HOST is set
		config = models.DBConfig{
			Driver:   "mysql",
			Host:     dbHost,
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
		}
	} else {
		// Default to SQLite
		log.Println("Using SQLite database configuration")
		config = models.DBConfig{
			Driver: "sqlite3",
			Name:   "local.db",
		}
	}

	log.Println("Initializing database connection...")
	db, err := models.InitDB(config)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	log.Println("Database initialized successfully")

	log.Println("Setting up document store...")
	store := models.NewDBDocumentStore(db)

	log.Println("Configuring router...")
	router := mux.NewRouter()

	// API routes
	log.Println("Registering API routes...")
	router.HandleFunc("/api/documents/signatures/request", func(w http.ResponseWriter, r *http.Request) {
		handlers.SignRequestHandler(w, r, store)
	}).Methods(http.MethodPost)

	router.HandleFunc("/api/documents/signatures/{request_id}/status", func(w http.ResponseWriter, r *http.Request) {
		handlers.SignatureStatusHandler(w, r, store)
	}).Methods(http.MethodGet)

	router.HandleFunc("/api/documents/signatures/{request_id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.DeleteSignatureHandler(w, r, store)
	}).Methods(http.MethodDelete)

	// Web routes
	deviceEntryHandler := handlers.NewDeviceEntryHandler()
	documentsHandler := handlers.NewDocumentsHandler(store)
	signatureHandler := handlers.NewSignatureHandler(store)

	// Register the documents handler routes
	router.HandleFunc("/documents/{device_id}", documentsHandler.ListDocuments).Methods("GET")

	// Register signature handler routes
	router.HandleFunc("/documents/sign/{request_id}", signatureHandler.ShowSignaturePage).Methods("GET")
	router.HandleFunc("/documents/sign/{request_id}", signatureHandler.ProcessSignature).Methods("POST")

	// Register root handler routes
	router.HandleFunc("/", deviceEntryHandler.ShowForm).Methods("GET")
	router.HandleFunc("/", deviceEntryHandler.ProcessForm).Methods("POST")

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Attempting to start server on port %s...\n", port)
	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
