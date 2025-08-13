package client_db

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/booking"
	"github.com/mishkahtherapy/brain/core/domain/client"
	"github.com/mishkahtherapy/brain/core/ports"
)

type ClientRepository struct {
	db ports.SQLDatabase
}

var (
	ErrReadingClientBookings = errors.New("error reading client bookings")
	ErrReadingClient         = errors.New("error reading client")
)

func NewClientRepository(database ports.SQLDatabase) ports.ClientRepository {
	return &ClientRepository{
		db: database,
	}
}

func (r *ClientRepository) Create(client *client.Client) error {
	query := `
		INSERT INTO clients (id, name, whatsapp_number, timezone_offset, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(
		query,
		client.ID,
		client.Name,
		client.WhatsAppNumber,
		client.TimezoneOffset,
		client.CreatedAt,
		client.UpdatedAt,
	)
	return err
}

func (r *ClientRepository) FindByIDs(ids []domain.ClientID) ([]*client.Client, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(ids))
	values := make([]interface{}, len(ids))
	for i := range ids {
		placeholders[i] = "?"
		values[i] = ids[i]
	}
	placeholdersStr := strings.Join(placeholders, ",")

	query := `
		SELECT id, name, whatsapp_number, timezone_offset, created_at, updated_at
		FROM clients
		WHERE id IN (%s)
	`
	query = fmt.Sprintf(query, placeholdersStr)

	rows, err := r.db.Query(query, values...)
	if err != nil {
		slog.Error("error querying clients", "error", err, "ids", ids)
		return nil, err
	}
	defer rows.Close()

	var clients []*client.Client
	for rows.Next() {
		var client client.Client
		err := rows.Scan(
			&client.ID,
			&client.Name,
			&client.WhatsAppNumber,
			&client.TimezoneOffset,
			&client.CreatedAt,
			&client.UpdatedAt,
		)
		if err != nil {
			slog.Error("error scanning client", "error", err)
			return nil, ErrReadingClient
		}
		clients = append(clients, &client)
	}

	return clients, nil
}

func (r *ClientRepository) GetByWhatsAppNumber(whatsAppNumber domain.WhatsAppNumber) (*client.Client, error) {
	query := `
		SELECT id, name, whatsapp_number, timezone_offset, created_at, updated_at
		FROM clients
		WHERE whatsapp_number = ?
	`
	row := r.db.QueryRow(query, whatsAppNumber)

	var client client.Client
	err := row.Scan(
		&client.ID,
		&client.Name,
		&client.WhatsAppNumber,
		&client.TimezoneOffset,
		&client.CreatedAt,
		&client.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Get booking IDs for this client
	bookingIDs, err := r.BulkGetClientBookings([]domain.ClientID{client.ID})
	if err != nil {
		return nil, err
	}

	client.Bookings = bookingIDs[client.ID]
	if client.Bookings == nil {
		client.Bookings = []booking.Booking{}
	}
	return &client, nil
}

func (r *ClientRepository) List() ([]*client.Client, error) {
	query := `
		SELECT id, name, whatsapp_number, timezone_offset, created_at, updated_at
		FROM clients
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []*client.Client
	for rows.Next() {
		var client client.Client
		err := rows.Scan(
			&client.ID,
			&client.Name,
			&client.WhatsAppNumber,
			&client.TimezoneOffset,
			&client.CreatedAt,
			&client.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Get booking IDs for this client
		bookingIDs, err := r.BulkGetClientBookings([]domain.ClientID{client.ID})
		if err != nil {
			return nil, err
		}

		client.Bookings = bookingIDs[client.ID]
		if client.Bookings == nil {
			client.Bookings = []booking.Booking{}
		}

		clients = append(clients, &client)
	}

	return clients, nil
}

func (r *ClientRepository) Update(client *client.Client) error {
	query := `
		UPDATE clients
		SET name = ?, whatsapp_number = ?, timezone_offset = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.Exec(
		query,
		client.Name,
		client.WhatsAppNumber,
		client.TimezoneOffset,
		client.UpdatedAt,
		client.ID,
	)
	return err
}

func (r *ClientRepository) Delete(id domain.ClientID) error {
	query := `DELETE FROM clients WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *ClientRepository) UpdateTimezoneOffset(id domain.ClientID, offsetMinutes domain.TimezoneOffset) error {
	query := `UPDATE clients SET timezone_offset = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, offsetMinutes, domain.NewUTCTimestamp(), id)
	return err
}

func (r *ClientRepository) BulkGetClientBookings(
	clientIDs []domain.ClientID,
) (map[domain.ClientID][]booking.Booking, error) {
	if len(clientIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(clientIDs))
	values := make([]interface{}, len(clientIDs))
	for i := range clientIDs {
		placeholders[i] = "?"
		values[i] = clientIDs[i]
	}
	placeholdersStr := strings.Join(placeholders, ",")
	query := `
		SELECT id, timeslot_id, therapist_id, start_time, duration_minutes, state
		FROM bookings
		WHERE client_id IN (%s)
		ORDER BY created_at DESC
	`
	query = fmt.Sprintf(query, placeholdersStr)

	rows, err := r.db.Query(query, values...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, ErrReadingClientBookings
	}
	defer rows.Close()

	bookings := make(map[domain.ClientID][]booking.Booking)
	for rows.Next() {
		var booking booking.Booking
		err := rows.Scan(
			&booking.ID,
			&booking.TimeSlotID,
			&booking.TherapistID,
			&booking.StartTime,
			&booking.Duration,
			&booking.State,
		)
		if err != nil {
			slog.Error("error scanning booking", "error", err)
			return nil, ErrReadingClientBookings
		}
		bookings[booking.ClientID] = append(bookings[booking.ClientID], booking)
	}

	return bookings, nil
}
