package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/mishkahtherapy/brain/adapters/api"
	bookingHandler "github.com/mishkahtherapy/brain/adapters/api/booking"
	clientHandler "github.com/mishkahtherapy/brain/adapters/api/client"
	scheduleHandler "github.com/mishkahtherapy/brain/adapters/api/schedule"
	specializationHandler "github.com/mishkahtherapy/brain/adapters/api/specialization"
	therapistHandler "github.com/mishkahtherapy/brain/adapters/api/therapist"
	timeslotHandler "github.com/mishkahtherapy/brain/adapters/api/timeslot"
	"github.com/mishkahtherapy/brain/adapters/db"
	"github.com/mishkahtherapy/brain/adapters/db/booking_db"
	"github.com/mishkahtherapy/brain/adapters/db/client_db"
	"github.com/mishkahtherapy/brain/adapters/db/session_db"
	"github.com/mishkahtherapy/brain/adapters/db/specialization_db"
	"github.com/mishkahtherapy/brain/adapters/db/therapist_db"
	"github.com/mishkahtherapy/brain/adapters/db/timeslot_db"
	"github.com/mishkahtherapy/brain/config"
	"github.com/mishkahtherapy/brain/core/usecases/booking/cancel_booking"
	"github.com/mishkahtherapy/brain/core/usecases/booking/confirm_booking"
	"github.com/mishkahtherapy/brain/core/usecases/booking/create_booking"
	"github.com/mishkahtherapy/brain/core/usecases/booking/get_booking"
	"github.com/mishkahtherapy/brain/core/usecases/booking/list_bookings_by_client"
	"github.com/mishkahtherapy/brain/core/usecases/booking/list_bookings_by_therapist"
	"github.com/mishkahtherapy/brain/core/usecases/booking/search_bookings"
	"github.com/mishkahtherapy/brain/core/usecases/client/create_client"
	"github.com/mishkahtherapy/brain/core/usecases/client/get_all_clients"
	"github.com/mishkahtherapy/brain/core/usecases/client/get_client"
	"github.com/mishkahtherapy/brain/core/usecases/schedule/get_schedule"
	"github.com/mishkahtherapy/brain/core/usecases/session/create_session"
	"github.com/mishkahtherapy/brain/core/usecases/session/get_meeting_link"
	"github.com/mishkahtherapy/brain/core/usecases/session/get_session"
	"github.com/mishkahtherapy/brain/core/usecases/session/list_sessions_admin"
	"github.com/mishkahtherapy/brain/core/usecases/session/list_sessions_by_client"
	"github.com/mishkahtherapy/brain/core/usecases/session/list_sessions_by_therapist"
	"github.com/mishkahtherapy/brain/core/usecases/session/update_meeting_url"
	"github.com/mishkahtherapy/brain/core/usecases/session/update_session_notes"
	"github.com/mishkahtherapy/brain/core/usecases/session/update_session_state"
	"github.com/mishkahtherapy/brain/core/usecases/specialization/get_all_specializations"
	"github.com/mishkahtherapy/brain/core/usecases/specialization/get_specialization"
	"github.com/mishkahtherapy/brain/core/usecases/specialization/new_specialization"
	"github.com/mishkahtherapy/brain/core/usecases/therapist/get_all_therapists"
	"github.com/mishkahtherapy/brain/core/usecases/therapist/get_therapist"
	"github.com/mishkahtherapy/brain/core/usecases/therapist/new_therapist"
	"github.com/mishkahtherapy/brain/core/usecases/therapist/update_therapist_device"
	"github.com/mishkahtherapy/brain/core/usecases/therapist/update_therapist_info"
	"github.com/mishkahtherapy/brain/core/usecases/therapist/update_therapist_specializations"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/bulk_toggle_therapist_timeslots"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/create_therapist_timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/delete_therapist_timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/get_therapist_timeslot"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/list_therapist_timeslots"
	"github.com/mishkahtherapy/brain/core/usecases/timeslot/update_therapist_timeslot"

	_ "github.com/glebarez/go-sqlite" // SQLite driver
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	err := config.LoadEnvFileIfExists(".env")
	if err != nil {
		slog.Error("Error loading env file", "error", err)
	}

	// Initialize database
	dbConfig := config.GetDBConfig()
	database := db.NewDatabase(dbConfig)
	defer database.Close()

	slog.Info("Database initialized successfully", slog.Group("db", "name", dbConfig.DBFilename, "schema", dbConfig.SchemaFile))

	// Initialize repositories
	specializationRepo := specialization_db.NewSpecializationRepository(database)
	therapistRepo := therapist_db.NewTherapistRepository(database)
	clientRepo := client_db.NewClientRepository(database)
	bookingRepo := booking_db.NewBookingRepository(database)
	sessionRepo := session_db.NewSessionRepository(database)
	timeSlotRepo := timeslot_db.NewTimeSlotRepository(database)

	// Initialize specialization usecases
	newSpecializationUsecase := new_specialization.NewUsecase(specializationRepo)
	getAllSpecializationsUsecase := get_all_specializations.NewUsecase(specializationRepo)
	getSpecializationUsecase := get_specialization.NewUsecase(specializationRepo)

	// Initialize therapist usecases
	newTherapistUsecase := new_therapist.NewUsecase(therapistRepo, specializationRepo)
	getAllTherapistsUsecase := get_all_therapists.NewUsecase(therapistRepo)
	getTherapistUsecase := get_therapist.NewUsecase(therapistRepo)
	updateTherapistInfoUsecase := update_therapist_info.NewUsecase(therapistRepo)
	updateTherapistSpecializationsUsecase := update_therapist_specializations.NewUsecase(therapistRepo, specializationRepo)
	updateTherapistDeviceUsecase := update_therapist_device.NewUsecase(therapistRepo)

	// Initialize timeslot usecases
	createTherapistTimeslotUsecase := create_therapist_timeslot.NewUsecase(therapistRepo, timeSlotRepo)
	getTherapistTimeslotUsecase := get_therapist_timeslot.NewUsecase(therapistRepo, timeSlotRepo)
	updateTherapistTimeslotUsecase := update_therapist_timeslot.NewUsecase(therapistRepo, timeSlotRepo)
	deleteTherapistTimeslotUsecase := delete_therapist_timeslot.NewUsecase(therapistRepo, timeSlotRepo)
	listTherapistTimeslotsUsecase := list_therapist_timeslots.NewUsecase(therapistRepo, timeSlotRepo)
	bulkToggleTherapistTimeslotsUsecase := bulk_toggle_therapist_timeslots.NewUsecase(therapistRepo, timeSlotRepo)

	// Initialize client usecases
	createClientUsecase := create_client.NewUsecase(clientRepo)
	getAllClientsUsecase := get_all_clients.NewUsecase(clientRepo)
	getClientUsecase := get_client.NewUsecase(clientRepo)

	// Initialize booking usecases
	createBookingUsecase := create_booking.NewUsecase(bookingRepo, therapistRepo, clientRepo, timeSlotRepo)
	getBookingUsecase := get_booking.NewUsecase(bookingRepo)
	confirmBookingUsecase := confirm_booking.NewUsecase(bookingRepo, sessionRepo)
	cancelBookingUsecase := cancel_booking.NewUsecase(bookingRepo)
	listBookingsByTherapistUsecase := list_bookings_by_therapist.NewUsecase(bookingRepo)
	listBookingsByClientUsecase := list_bookings_by_client.NewUsecase(bookingRepo)
	searchBookingsUsecase := search_bookings.NewUsecase(bookingRepo, therapistRepo, clientRepo)

	// Initialize session usecases
	createSessionUsecase := create_session.NewUsecase(sessionRepo, bookingRepo, therapistRepo, clientRepo, timeSlotRepo)
	getSessionUsecase := get_session.NewUsecase(sessionRepo)
	updateSessionStateUsecase := update_session_state.NewUsecase(sessionRepo)
	updateSessionNotesUsecase := update_session_notes.NewUsecase(sessionRepo)
	updateMeetingURLUsecase := update_meeting_url.NewUsecase(sessionRepo)
	listSessionsByTherapistUsecase := list_sessions_by_therapist.NewUsecase(sessionRepo)
	listSessionsByClientUsecase := list_sessions_by_client.NewUsecase(sessionRepo)
	listSessionsAdminUsecase := list_sessions_admin.NewUsecase(sessionRepo)
	getMeetingLinkUsecase := get_meeting_link.NewUsecase(sessionRepo)

	// Initialize schedule usecases
	getScheduleUsecase := get_schedule.NewUsecase(therapistRepo, timeSlotRepo, bookingRepo)

	// Initialize handlers
	specializationHandler := specializationHandler.NewSpecializationHandler(
		*newSpecializationUsecase,
		*getAllSpecializationsUsecase,
		*getSpecializationUsecase,
	)

	therapistHandler := therapistHandler.NewTherapistHandler(
		*newTherapistUsecase,
		*getAllTherapistsUsecase,
		*getTherapistUsecase,
		*updateTherapistInfoUsecase,
		*updateTherapistSpecializationsUsecase,
		*updateTherapistDeviceUsecase,
	)

	clientHandler := clientHandler.NewClientHandler(
		*createClientUsecase,
		*getAllClientsUsecase,
		*getClientUsecase,
	)

	bookingHandler := bookingHandler.NewBookingHandler(
		*createBookingUsecase,
		*getBookingUsecase,
		*confirmBookingUsecase,
		*cancelBookingUsecase,
		*listBookingsByTherapistUsecase,
		*listBookingsByClientUsecase,
		*searchBookingsUsecase,
	)

	sessionHandler := api.NewSessionHandler(
		*createSessionUsecase,
		*getSessionUsecase,
		*updateSessionStateUsecase,
		*updateSessionNotesUsecase,
		*updateMeetingURLUsecase,
		*listSessionsByTherapistUsecase,
		*listSessionsByClientUsecase,
		*listSessionsAdminUsecase,
	)

	meetingLinkProxyHandler := api.NewMeetingLinkProxyHandler(
		*getMeetingLinkUsecase,
	)

	scheduleHandler := scheduleHandler.NewScheduleHandler(
		*getScheduleUsecase,
	)

	timeslotHandler := timeslotHandler.NewTimeslotHandler(
		bulkToggleTherapistTimeslotsUsecase,
		*createTherapistTimeslotUsecase,
		*getTherapistTimeslotUsecase,
		*updateTherapistTimeslotUsecase,
		*deleteTherapistTimeslotUsecase,
		*listTherapistTimeslotsUsecase,
	)

	// Setup HTTP routes
	mux := http.NewServeMux()

	// Register specialization routes
	specializationHandler.RegisterRoutes(mux)

	// Register therapist routes
	therapistHandler.RegisterRoutes(mux)

	// Register client routes
	clientHandler.RegisterRoutes(mux)

	// Register booking routes
	bookingHandler.RegisterRoutes(mux)

	// Register session routes
	sessionHandler.RegisterRoutes(mux)

	// Register meeting link proxy routes
	meetingLinkProxyHandler.RegisterRoutes(mux)

	// Register schedule routes
	scheduleHandler.RegisterRoutes(mux)

	// Register timeslot routes
	timeslotHandler.RegisterRoutes(mux)

	// Add health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"therapist-api"}`))
	})

	var middleWareStack []func(http.Handler) http.Handler
	var handler http.Handler
	if config.IsDevelopment() {
		// Add CORS middleware
		middleWareStack = append(middleWareStack, corsMiddleware)
	}

	handler = loggingMiddleware(mux)
	for _, middleware := range middleWareStack {
		handler = middleware(handler)
	}

	// Start server
	port := getEnvOrDefault("PORT", "8090")
	slog.Info("Starting server", "port", port)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// loggingMiddleware logs the HTTP method, path, status code, and response time for each request.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap ResponseWriter to capture status
		rw := &statusCapturingResponseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		slog.Info(
			"HTTP",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration", duration.String(),
			"user_agent", r.UserAgent(),
			"remote_addr", r.RemoteAddr,
		)
	})
}

// statusCapturingResponseWriter wraps http.ResponseWriter to capture the status code
// so it can be logged after the handler completes.
type statusCapturingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusCapturingResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
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
