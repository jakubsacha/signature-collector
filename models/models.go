package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// DocumentSection represents a section in the document content
type DocumentSection struct {
	ID               string  `json:"id"`
	Type             string  `json:"type"`
	Content          string  `json:"content"`
	ConsentType      *string `json:"consent_type,omitempty"`
	ConsentGranted   *bool   `json:"consent_granted,omitempty"`
	ConsentMandatory *bool   `json:"consent_mandatory,omitempty"`
	ConsentDefault   *bool   `json:"consent_default,omitempty"`
}

// Document represents a document to be signed
type Document struct {
	ID              string            `json:"id"`
	DocumentTitle   string            `json:"document_title"`
	DocumentContent []DocumentSection `json:"document_content"`
	SignerName      string            `json:"signer_name"`
	SignerEmail     string            `json:"signer_email"`
	DeviceID        string            `json:"device_id"`
	CallbackURL     string            `json:"callback_url"`
	Status          string            `json:"status"`
}

// DBConfig holds the configuration for the database connection
type DBConfig struct {
	Driver   string
	User     string
	Password string
	Name     string
	Host     string
}

// InitDB initializes the database connection based on the provided configuration
func InitDB(config DBConfig) (*sql.DB, error) {
	var dsn string
	if config.Driver == "mysql" {
		dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s", config.User, config.Password, config.Host, config.Name)
	} else if config.Driver == "sqlite3" {
		dsn = config.Name
	} else {
		return nil, fmt.Errorf("unsupported driver: %s", config.Driver)
	}

	DB, err := sql.Open(config.Driver, dsn)
	if err != nil {
		return nil, err
	}

	return DB, DB.Ping()
}

// DocumentStore defines the interface for document operations
// This allows for mocking in tests.
type DocumentStore interface {
	AddDocument(doc Document) (string, error)
	ListDocuments(deviceID string) ([]Document, error)
	UpdateDocumentStatus(requestID, status string) error
	GetSignatureStatus(requestID string) (string, string, error)
	GetDocument(requestID string) (Document, error)
	UpdateDocumentSignature(requestID string, signatureData string) error
	StoreConsents(requestID string, consents []Consent) error
}

func NewDBDocumentStore(db *sql.DB) DocumentStore {
	return &DBDocumentStore{db: db}
}

// DefaultDocumentStore is the default implementation of DocumentStore
// It uses the global DB connection.
type DBDocumentStore struct {
	db *sql.DB
}

func (ds DBDocumentStore) AddDocument(doc Document) (string, error) {
	documentContent, err := json.Marshal(doc.DocumentContent)
	if err != nil {
		return "", fmt.Errorf("error marshaling document content: %v", err)
	}

	// generate UUID
	uuid := uuid.NewString()

	query := "INSERT INTO documents (id, document_title, document_content, signer_name, signer_email, device_id, callback_url, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	_, err = ds.db.Exec(query, uuid, doc.DocumentTitle, documentContent, doc.SignerName, doc.SignerEmail, doc.DeviceID, doc.CallbackURL, doc.Status)
	if err != nil {
		return "", fmt.Errorf("error inserting document: %v", err)
	}

	return uuid, nil
}

// ListDocuments lists all pending documents for a specific device
func (ds DBDocumentStore) ListDocuments(deviceID string) ([]Document, error) {
	query := `
		SELECT id, document_title, document_content, signer_name, signer_email, device_id, callback_url, status 
		FROM documents 
		WHERE device_id = ? AND status = 'pending'
		ORDER BY created_at DESC`

	rows, err := ds.db.Query(query, deviceID)
	if err != nil {
		return nil, fmt.Errorf("error querying documents: %v", err)
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var doc Document
		var documentContent []byte
		if err := rows.Scan(&doc.ID, &doc.DocumentTitle, &documentContent, &doc.SignerName, &doc.SignerEmail, &doc.DeviceID, &doc.CallbackURL, &doc.Status); err != nil {
			return nil, fmt.Errorf("error scanning document: %v", err)
		}
		if err := json.Unmarshal(documentContent, &doc.DocumentContent); err != nil {
			return nil, fmt.Errorf("error unmarshaling document content: %v", err)
		}
		documents = append(documents, doc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return documents, nil
}

// UpdateDocumentStatus updates the status of a document
func (ds DBDocumentStore) UpdateDocumentStatus(requestID, status string) error {
	query := "UPDATE documents SET status = ? WHERE id = ?"
	_, err := ds.db.Exec(query, status, requestID)
	return err
}

// GetSignatureStatus retrieves the status and signed document URL for a document
func (ds DBDocumentStore) GetSignatureStatus(requestID string) (string, string, error) {
	query := "SELECT status, document_content FROM documents WHERE id = ?"
	var status string
	var documentContent []byte
	err := ds.db.QueryRow(query, requestID).Scan(&status, &documentContent)
	if err != nil {
		return "", "", err
	}

	var content []DocumentSection
	if err := json.Unmarshal(documentContent, &content); err != nil {
		return "", "", err
	}

	// Find the first section with type "text" to use as document URL
	var documentURL string
	for _, section := range content {
		if section.Type == "text" {
			documentURL = section.Content
			break
		}
	}

	return status, documentURL, nil
}

// GetDocument retrieves a document by its ID
func (ds DBDocumentStore) GetDocument(requestID string) (Document, error) {
	query := `
		SELECT id, document_title, document_content, signer_name, signer_email, device_id, callback_url, status 
		FROM documents 
		WHERE id = ?`

	var doc Document
	var documentContent []byte
	err := ds.db.QueryRow(query, requestID).Scan(
		&doc.ID,
		&doc.DocumentTitle,
		&documentContent,
		&doc.SignerName,
		&doc.SignerEmail,
		&doc.DeviceID,
		&doc.CallbackURL,
		&doc.Status,
	)
	if err != nil {
		return Document{}, err
	}

	if err := json.Unmarshal(documentContent, &doc.DocumentContent); err != nil {
		return Document{}, fmt.Errorf("error unmarshaling document content: %v", err)
	}

	return doc, nil
}

// UpdateDocumentSignature updates the signature data for a document
func (ds DBDocumentStore) UpdateDocumentSignature(requestID string, signatureData string) error {
	query := "UPDATE documents SET signature_data = ? WHERE id = ?"
	_, err := ds.db.Exec(query, signatureData, requestID)
	return err
}

// Consent represents a user's consent
type Consent struct {
	ConsentType string    `json:"consent_type"`
	Granted     bool      `json:"granted"`
	Timestamp   time.Time `json:"timestamp"`
}

// StoreConsents stores the consents for a document
func (ds DBDocumentStore) StoreConsents(requestID string, consents []Consent) error {
	// Convert consents to JSON for storage
	consentsJSON, err := json.Marshal(consents)
	if err != nil {
		return fmt.Errorf("error marshaling consents: %v", err)
	}

	query := "UPDATE documents SET consents = ? WHERE id = ?"
	_, err = ds.db.Exec(query, consentsJSON, requestID)
	return err
}

// InMemoryDocumentStore is an in-memory implementation of the DocumentStore interface
// for testing purposes.
type InMemoryDocumentStore struct {
	documents map[string]Document
}

func NewInMemoryDocumentStore() *InMemoryDocumentStore {
	return &InMemoryDocumentStore{
		documents: make(map[string]Document),
	}
}

func (m *InMemoryDocumentStore) AddDocument(doc Document) (string, error) {
	id := uuid.NewString()
	doc.ID = id
	m.documents[id] = doc
	return id, nil
}

func (m *InMemoryDocumentStore) ListDocuments(deviceID string) ([]Document, error) {
	var result []Document
	for _, doc := range m.documents {
		if doc.DeviceID == deviceID && doc.Status == "pending" {
			result = append(result, doc)
		}
	}
	// Sort documents by ID to maintain consistent order
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result, nil
}

func (m *InMemoryDocumentStore) UpdateDocumentStatus(requestID, status string) error {
	doc, exists := m.documents[requestID]
	if !exists {
		return fmt.Errorf("document not found")
	}
	doc.Status = status
	m.documents[requestID] = doc
	return nil
}

func (m *InMemoryDocumentStore) GetSignatureStatus(requestID string) (string, string, error) {
	doc, exists := m.documents[requestID]
	if !exists {
		return "", "", fmt.Errorf("document not found")
	}

	// Find the first section with type "text" to use as document URL
	var documentURL string
	for _, section := range doc.DocumentContent {
		if section.Type == "text" {
			documentURL = section.Content
			break
		}
	}

	return doc.Status, documentURL, nil
}

func (m *InMemoryDocumentStore) GetDocument(requestID string) (Document, error) {
	doc, exists := m.documents[requestID]
	if !exists {
		return Document{}, fmt.Errorf("document not found")
	}
	return doc, nil
}

func (m *InMemoryDocumentStore) UpdateDocumentSignature(requestID string, signatureData string) error {
	doc, exists := m.documents[requestID]
	if !exists {
		return fmt.Errorf("document not found")
	}
	// In a real implementation, we would store the signature data in a proper field
	// For now, we'll just update the status
	doc.Status = "completed"
	m.documents[requestID] = doc
	return nil
}

func (m *InMemoryDocumentStore) StoreConsents(requestID string, consents []Consent) error {
	_, exists := m.documents[requestID]
	if !exists {
		return fmt.Errorf("document not found")
	}
	// In memory implementation doesn't actually store consents
	return nil
}
