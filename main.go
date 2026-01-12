package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jakubsacha/signature-collector/handlers"
	"github.com/jakubsacha/signature-collector/i18n"
	"github.com/jakubsacha/signature-collector/logging"
	"github.com/jakubsacha/signature-collector/models"
	"github.com/joho/godotenv"
)

func main() {
	// Set up logging
	logging.Info("Starting application...")

	// Load .env file
	logging.Info("Loading .env file...")
	err := godotenv.Load()
	if err != nil {
		logging.Info("Didn't load .env file")
	}
	logging.Info(".env file loaded successfully")

	// check if API_USER and API_PASS are set
	if os.Getenv("BASEAUTH_USER") == "" || os.Getenv("BASEAUTH_PASS") == "" {
		logging.Fatal("BASEAUTH_USER and BASEAUTH_PASS must be set")
	}

	// check if API_TOKEN is set
	if os.Getenv("API_TOKEN") == "" {
		logging.Fatal("API_TOKEN must be set")
	}

	// Initialize i18n
	logging.Info("Initializing i18n...")
	err = i18n.Init(os.Getenv("LANGUAGE"))
	if err != nil {
		logging.WithField("error", err.Error()).Fatal("Error initializing i18n")
	}
	logging.Info("i18n initialized successfully")

	// Initialize the database
	logging.Info("Setting up database configuration...")
	var config models.DBConfig
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		logging.WithField("driver", "mysql").Info("Using MySQL database configuration")
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
		logging.WithField("driver", "sqlite3").Info("Using SQLite database configuration")
		config = models.DBConfig{
			Driver: "sqlite3",
			Name:   "local.db",
		}
	}

	logging.Info("Initializing database connection...")
	db, err := models.InitDB(config)
	if err != nil {
		logging.WithField("error", err.Error()).Fatal("Error initializing database")
	}
	logging.Info("Database initialized successfully")

	logging.Info("Setting up document store...")
	store := models.NewDBDocumentStore(db)

	logging.Info("Configuring router...")
	router := mux.NewRouter()

	// API routes with token authentication
	router.HandleFunc("/api/documents/signatures/request", tokenAuth(func(w http.ResponseWriter, r *http.Request) {
		handlers.SignRequestHandler(w, r, store)
	})).Methods(http.MethodPost)

	router.HandleFunc("/api/documents/signatures/{request_id}/status", tokenAuth(func(w http.ResponseWriter, r *http.Request) {
		handlers.SignatureStatusHandler(w, r, store)
	})).Methods(http.MethodGet)

	router.HandleFunc("/api/documents/signatures/{request_id}", tokenAuth(func(w http.ResponseWriter, r *http.Request) {
		handlers.DeleteSignatureHandler(w, r, store)
	})).Methods(http.MethodDelete)

	// Web routes with basic authentication
	deviceEntryHandler := handlers.NewDeviceEntryHandler()
	documentsHandler := handlers.NewDocumentsHandler(store)
	signatureHandler := handlers.NewSignatureHandler(store)

	// Register the documents handler routes
	router.HandleFunc("/documents/{device_id}", basicAuth(documentsHandler.ListDocuments)).Methods("GET")

	// Register signature handler routes
	router.HandleFunc("/documents/sign/{request_id}", basicAuth(signatureHandler.ShowSignaturePage)).Methods("GET")
	router.HandleFunc("/documents/sign/{request_id}", basicAuth(signatureHandler.ProcessSignature)).Methods("POST")

	// Register root handler routes
	router.HandleFunc("/", basicAuth(deviceEntryHandler.ShowForm)).Methods("GET")
	router.HandleFunc("/", basicAuth(deviceEntryHandler.ProcessForm)).Methods("POST")

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logging.WithField("port", port).Info("Attempting to start server")
	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		logging.WithField("error", err.Error()).Fatal("Error starting server")
	}
}

// basicAuth is a middleware for basic authentication
func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || !validateBasicAuth(user, pass) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// validateBasicAuth checks the provided username and password
func validateBasicAuth(user, pass string) bool {
	expectedUser := os.Getenv("BASEAUTH_USER")
	expectedPass := os.Getenv("BASEAUTH_PASS")
	return user == expectedUser && pass == expectedPass
}

// tokenAuth is a middleware for token-based authentication
func tokenAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		if !validateToken(token) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// validateToken checks the provided token
func validateToken(token string) bool {
	expectedToken := os.Getenv("API_TOKEN")
	return token == "Bearer "+expectedToken
}
