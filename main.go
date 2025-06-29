package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/mishkahtherapy/brain/adapters/api"
	"github.com/mishkahtherapy/brain/adapters/db"
	"github.com/mishkahtherapy/brain/adapters/db/specialization"
	"github.com/mishkahtherapy/brain/adapters/db/therapist"
	"github.com/mishkahtherapy/brain/core/usecases/specialization/get_all_specializations"
	"github.com/mishkahtherapy/brain/core/usecases/specialization/get_specialization"
	"github.com/mishkahtherapy/brain/core/usecases/specialization/new_specialization"
	"github.com/mishkahtherapy/brain/core/usecases/therapist/get_all_therapists"
	"github.com/mishkahtherapy/brain/core/usecases/therapist/get_therapist"
	"github.com/mishkahtherapy/brain/core/usecases/therapist/new_therapist"
	"github.com/mishkahtherapy/brain/core/usecases/therapist/update_therapist_specializations"

	_ "github.com/glebarez/go-sqlite" // SQLite driver
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Initialize database
	dbConfig := db.DatabaseConfig{
		DBFilename: "./brain.db", // SQLite database file
	}
	database := db.NewDatabase(dbConfig)
	defer database.Close()

	slog.Info("Database initialized successfully")

	// Initialize repositories
	specializationRepo := specialization.NewSpecializationRepository(database)
	therapistRepo := therapist.NewTherapistRepository(database)

	// Initialize specialization usecases
	newSpecializationUsecase := new_specialization.NewUsecase(specializationRepo)
	getAllSpecializationsUsecase := get_all_specializations.NewUsecase(specializationRepo)
	getSpecializationUsecase := get_specialization.NewUsecase(specializationRepo)

	// Initialize therapist usecases
	newTherapistUsecase := new_therapist.NewUsecase(therapistRepo)
	getAllTherapistsUsecase := get_all_therapists.NewUsecase(therapistRepo)
	getTherapistUsecase := get_therapist.NewUsecase(therapistRepo)
	updateTherapistSpecializationsUsecase := update_therapist_specializations.NewUsecase(therapistRepo, specializationRepo)

	// Initialize handlers
	specializationHandler := api.NewSpecializationHandler(
		*newSpecializationUsecase,
		*getAllSpecializationsUsecase,
		*getSpecializationUsecase,
	)

	therapistHandler := api.NewTherapistHandler(
		*newTherapistUsecase,
		*getAllTherapistsUsecase,
		*getTherapistUsecase,
		*updateTherapistSpecializationsUsecase,
	)

	// Setup HTTP routes
	mux := http.NewServeMux()

	// Register specialization routes
	specializationHandler.RegisterRoutes(mux)

	// Register therapist routes
	therapistHandler.RegisterRoutes(mux)

	// Add health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"therapist-api"}`))
	})

	// Add CORS middleware
	handler := corsMiddleware(mux)

	// Start server
	port := getEnvOrDefault("PORT", "8080")
	slog.Info("Starting server", "port", port)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// corsMiddleware adds CORS headers to allow cross-origin requests
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
