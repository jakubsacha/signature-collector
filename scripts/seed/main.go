package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jakubsacha/signature-collector/models"
	"github.com/joho/godotenv"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Initialize the database
	config := models.DBConfig{
		Driver: "sqlite3",
		Name:   "local.db",
	}

	db, err := models.InitDB(config)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	// Truncate the documents table
	if _, err := db.Exec("DELETE FROM documents"); err != nil {
		log.Fatalf("Error truncating documents table: %v", err)
	}
	log.Println("Documents table truncated successfully")

	// Create sample documents
	store := models.NewDBDocumentStore(db)

	// Sample device IDs
	deviceIDs := []string{"tablet1", "tablet2", "tablet3"}

	// Sample document sections
	sampleSections := []models.DocumentSection{
		{
			ID:      "header",
			Type:    "text",
			Content: "Zgoda na przetwarzanie danych osobowych",
		},
		{
			ID:      "personal_info",
			Type:    "text",
			Content: "Imie nazwisko\nAdres\n05-123 Warszawa\n\nData: 2024-12-27",
		},
		{
			ID:      "data_list",
			Type:    "text",
			Content: "1. Ja niżej podpisany/a wyrażam zgodnę na przetwarzanie następująych danych osobowych:\n(1) imię i nazwisko,\n(2) numer ewidencyjny PESEL,\n(3) adres e-mail,\n(4) adres IP,\n(5) numery telefonów,\n(6) adres zamieszkania,\n(7) NIP,",
		},
		{
			ID:      "administrator_info",
			Type:    "text",
			Content: "2. Zostałem poinformowany/a, że Administatorem moich danych osobowych jest\nWETERYNARZ POD ŁAPĄ SPÓŁKA Z OGRANICZONĄ ODPOWIEDZIALNOŚCIĄ\nChodecka 14\n03-350 Warszawa\nMoje dane osobowe będą przez niego przechowywane i przetwarzane w celach bezpośrednio związanych z usługami świadczonymi przez WETERYNARZ POD ŁAPĄ SPÓŁKA Z OGRANICZONĄ ODPOWIEDZIALNOŚCIĄ, przez czas nieokreślony, od czasu zgłoszenia przeze mnie sprzeciwu, dotyczącego ich dalszego przetwarzania, co skutkować będzie ich usunięciem.",
		},
		{
			ID:      "gdpr_info",
			Type:    "text",
			Content: "3. Przetwarzanie danych osobowych odbywać sie będzie zgodnie z warunkami określonymi w rozporządzeniau Parlamentu Europejskiego i Rady (UE) 2016/679 z 27.04.2016 r. w sprawie ochrony osób fizycznych w związku z przetwarzaniem danych osobowych i w sprawie swobodnego przeplywu takich danych oraz uchylenia dyrektywy 95/46/WE (ogólne rozporządzenie o ochdonie danych) (Dz.Urz. UE L 119, s. 1) - dalej RODO.",
		},
		{
			ID:               "consent_database",
			Type:             "consent",
			Content:          "Wyrażam zgodę na umieszczenie moich danych adresowych oraz numeru telefonu i adresu e-mail w bazie danych",
			ConsentType:      strPtr("database_storage"),
			ConsentGranted:   boolPtr(false),
			ConsentMandatory: boolPtr(true),
			ConsentDefault:   boolPtr(true),
		},
		{
			ID:               "consent_sms",
			Type:             "consent",
			Content:          "Wyrażam zgodę na otrzymywanie powiadomień związanych z leczeniem zwierzęcia lub zabiegami profilaktycznymi na podany numer telefonu",
			ConsentType:      strPtr("sms_notifications"),
			ConsentGranted:   boolPtr(false),
			ConsentMandatory: boolPtr(false),
			ConsentDefault:   boolPtr(true),
		},
		{
			ID:               "consent_email",
			Type:             "consent",
			Content:          "Wyrażam zgodę na otrzymywanie wiadomości e-mail związanych z leczeniem zwierzęcia na podany adres e-mail",
			ConsentType:      strPtr("email_notifications"),
			ConsentGranted:   boolPtr(false),
			ConsentMandatory: boolPtr(false),
			ConsentDefault:   boolPtr(true),
		},
		{
			ID:               "consent_lab_results",
			Type:             "consent",
			Content:          "Wyrażam zgodę na przesłanie na podany przeze mnie adres mailowy wyników mojego zwierzęcia przez laboratorium (Vetlab, Labwet, Laboklin, Alab i inne)",
			ConsentType:      strPtr("lab_results"),
			ConsentGranted:   boolPtr(false),
			ConsentMandatory: boolPtr(false),
			ConsentDefault:   boolPtr(true),
		},
		{
			ID:               "consent_safe_animal",
			Type:             "consent",
			Content:          "Wyrażam zgodę na przekazanie moich danych osobowych do bazy danych SAFE-ANIMAL",
			ConsentType:      strPtr("safe_animal"),
			ConsentGranted:   boolPtr(false),
			ConsentMandatory: boolPtr(false),
			ConsentDefault:   boolPtr(true),
		},
		{
			ID:      "rights_info",
			Type:    "text",
			Content: "5. Rozumiem, że zgodnie z RODO przysługuje mi również prawdo do:\na) dostępu do swoich danych oraz otrzymania ich kopii;\nb) sprostowania (poprawiania) swoich danych;\nc) usunięcia danych, ograniczenia przetwarzania danych;\nd) przenoszenia danych;\ne) wniesienia skargi do organu nadzorczego.",
		},
	}

	// Create multiple documents for each device
	for i, deviceID := range deviceIDs {
		log.Printf("Creating documents for device: %s", deviceID)

		// Create 3 documents per device with different statuses
		statuses := []string{"pending", "completed", "pending"}

		for j, status := range statuses {
			doc := models.Document{
				ID:              uuid.New().String(),
				DocumentContent: sampleSections,
				DocumentTitle:   "Zgoda na przetwarzanie danych osobowych",
				SignerName:      fmt.Sprintf("User %d-%d", i+1, j+1),
				SignerEmail:     fmt.Sprintf("user%d-%d@example.com", i+1, j+1),
				DeviceID:        deviceID,
				CallbackURL:     "https://example.com/callback",
				Status:          status,
			}

			id, err := store.AddDocument(doc)
			if err != nil {
				log.Printf("Error adding document: %v", err)
				continue
			}

			docJSON, _ := json.MarshalIndent(doc, "", "  ")
			log.Printf("Added document %s:\n%s\n", id, docJSON)
		}
	}

	// Verify the data was inserted
	log.Println("\nVerifying inserted data:")
	for _, deviceID := range deviceIDs {
		docs, err := store.ListDocuments(deviceID)
		if err != nil {
			log.Printf("Error listing documents for device %s: %v", deviceID, err)
			continue
		}
		log.Printf("Device %s has %d documents", deviceID, len(docs))
		for _, doc := range docs {
			log.Printf("  - Document %s: Status: %s, Signer: %s", doc.ID, doc.Status, doc.SignerName)
		}
	}

	log.Println("\nDatabase seeded successfully!")
}

func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
