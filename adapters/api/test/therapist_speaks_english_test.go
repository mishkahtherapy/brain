package test

import (
	"testing"

	"github.com/mishkahtherapy/brain/adapters/db/specialization"
	"github.com/mishkahtherapy/brain/adapters/db/therapist"
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/usecases/therapist/get_therapist"
	"github.com/mishkahtherapy/brain/core/usecases/therapist/new_therapist"

	_ "github.com/glebarez/go-sqlite"
)

func TestTherapistSpeaksEnglish(t *testing.T) {
	// Setup test database
	dbInstance, cleanup := setupTherapistTestDB(t)
	defer cleanup()

	// Setup repository
	therapistRepo := therapist.NewTherapistRepository(dbInstance)
	specializationRepo := specialization.NewSpecializationRepository(dbInstance)

	// Setup usecases
	createTherapistUsecase := new_therapist.NewUsecase(therapistRepo, specializationRepo)
	getTherapistUsecase := get_therapist.NewUsecase(therapistRepo)

	t.Run("Create therapist with speaksEnglish=true", func(t *testing.T) {
		// Create therapist with speaksEnglish=true
		input := new_therapist.Input{
			Name:           "English Speaking Therapist",
			Email:          "english@example.com",
			PhoneNumber:    "+1234567890",
			WhatsAppNumber: "+1234567890",
			SpeaksEnglish:  true,
		}

		therapist, err := createTherapistUsecase.Execute(input)
		if err != nil {
			t.Fatalf("Failed to create therapist: %v", err)
		}
		if !therapist.SpeaksEnglish {
			t.Error("Expected SpeaksEnglish to be true, but it was false")
		}

		// Retrieve therapist and verify speaksEnglish field
		retrieved, err := getTherapistUsecase.Execute(therapist.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve therapist: %v", err)
		}
		if !retrieved.SpeaksEnglish {
			t.Error("Expected retrieved therapist SpeaksEnglish to be true, but it was false")
		}
	})

	t.Run("Create therapist with speaksEnglish=false", func(t *testing.T) {
		// Create therapist with speaksEnglish=false
		input := new_therapist.Input{
			Name:           "Non-English Speaking Therapist",
			Email:          "non-english@example.com",
			PhoneNumber:    "+1234567891",
			WhatsAppNumber: "+1234567891",
			SpeaksEnglish:  false,
		}

		therapist, err := createTherapistUsecase.Execute(input)
		if err != nil {
			t.Fatalf("Failed to create therapist: %v", err)
		}
		if therapist.SpeaksEnglish {
			t.Error("Expected SpeaksEnglish to be false, but it was true")
		}

		// Retrieve therapist and verify speaksEnglish field
		retrieved, err := getTherapistUsecase.Execute(therapist.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve therapist: %v", err)
		}
		if retrieved.SpeaksEnglish {
			t.Error("Expected retrieved therapist SpeaksEnglish to be false, but it was true")
		}
	})

	t.Run("Default value for speaksEnglish should be false", func(t *testing.T) {
		// This test verifies the schema default is working
		// First we'll create a therapist manually without setting speaksEnglish
		// to simulate the default behavior

		therapistID := domain.NewTherapistID()
		timestamp := domain.NewUTCTimestamp()

		// Insert directly into the database without setting speaksEnglish
		query := `
			INSERT INTO therapists (id, name, email, phone_number, whatsapp_number, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`
		_, err := dbInstance.Exec(
			query,
			therapistID,
			"Default Therapist",
			"default@example.com",
			"+1234567892",
			"+1234567892",
			timestamp,
			timestamp,
		)
		if err != nil {
			t.Fatalf("Failed to insert therapist directly: %v", err)
		}

		// Retrieve therapist and verify speaksEnglish defaults to false
		retrieved, err := getTherapistUsecase.Execute(therapistID)
		if err != nil {
			t.Fatalf("Failed to retrieve therapist: %v", err)
		}
		if retrieved.SpeaksEnglish {
			t.Error("Expected default SpeaksEnglish to be false, but it was true")
		}
	})

	t.Run("Update should preserve speaksEnglish value", func(t *testing.T) {
		// First create a therapist with speaksEnglish=true
		input := new_therapist.Input{
			Name:           "Updatable Therapist",
			Email:          "updatable@example.com",
			PhoneNumber:    "+1234567893",
			WhatsAppNumber: "+1234567893",
			SpeaksEnglish:  true,
		}

		therapist, err := createTherapistUsecase.Execute(input)
		if err != nil {
			t.Fatalf("Failed to create therapist: %v", err)
		}
		if !therapist.SpeaksEnglish {
			t.Error("Expected SpeaksEnglish to be true, but it was false")
		}

		// Update some non-speaksEnglish fields in the database
		updateQuery := `
			UPDATE therapists
			SET name = ?, phone_number = ?
			WHERE id = ?
		`
		_, err = dbInstance.Exec(
			updateQuery,
			"Updated Therapist Name",
			"+9876543210",
			therapist.ID,
		)
		if err != nil {
			t.Fatalf("Failed to update therapist: %v", err)
		}

		// Retrieve therapist and verify speaksEnglish is still true
		retrieved, err := getTherapistUsecase.Execute(therapist.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve therapist: %v", err)
		}
		if !retrieved.SpeaksEnglish {
			t.Error("Expected SpeaksEnglish to still be true after update, but it was false")
		}
		if retrieved.Name != "Updated Therapist Name" {
			t.Errorf("Expected name to be %q, got %q", "Updated Therapist Name", retrieved.Name)
		}
		if retrieved.PhoneNumber != domain.PhoneNumber("+9876543210") {
			t.Errorf("Expected phone number to be %q, got %q", "+9876543210", retrieved.PhoneNumber)
		}
	})
}

// The setupTherapistTestDB function is already declared in therapist_e2e_test.go
// We're in the same package, so we can use it directly without redeclaring it
