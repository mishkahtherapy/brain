package client

import (
	"database/sql"

	"github.com/mishkahtherapy/brain/adapters/db"
	"github.com/mishkahtherapy/brain/core/domain"
)

type ClientRepository struct {
	db db.SQLDatabase
}

func NewClientRepository(database db.SQLDatabase) *ClientRepository {
	return &ClientRepository{
		db: database,
	}
}

func (r *ClientRepository) Create(client *domain.Client) error {
	query := `
		INSERT INTO clients (id, name, whatsapp_number, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(
		query,
		client.ID,
		client.Name,
		client.WhatsAppNumber,
		client.CreatedAt,
		client.UpdatedAt,
	)
	return err
}

func (r *ClientRepository) GetByID(id domain.ClientID) (*domain.Client, error) {
	query := `
		SELECT id, name, whatsapp_number, created_at, updated_at
		FROM clients
		WHERE id = ?
	`
	row := r.db.QueryRow(query, id)

	var client domain.Client
	err := row.Scan(
		&client.ID,
		&client.Name,
		&client.WhatsAppNumber,
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
	bookingIDs, err := r.getBookingIDsForClient(client.ID)
	if err != nil {
		return nil, err
	}
	client.BookingIDs = bookingIDs

	return &client, nil
}

func (r *ClientRepository) GetByWhatsAppNumber(whatsAppNumber domain.WhatsAppNumber) (*domain.Client, error) {
	query := `
		SELECT id, name, whatsapp_number, created_at, updated_at
		FROM clients
		WHERE whatsapp_number = ?
	`
	row := r.db.QueryRow(query, whatsAppNumber)

	var client domain.Client
	err := row.Scan(
		&client.ID,
		&client.Name,
		&client.WhatsAppNumber,
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
	bookingIDs, err := r.getBookingIDsForClient(client.ID)
	if err != nil {
		return nil, err
	}
	client.BookingIDs = bookingIDs

	return &client, nil
}

func (r *ClientRepository) List() ([]*domain.Client, error) {
	query := `
		SELECT id, name, whatsapp_number, created_at, updated_at
		FROM clients
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []*domain.Client
	for rows.Next() {
		var client domain.Client
		err := rows.Scan(
			&client.ID,
			&client.Name,
			&client.WhatsAppNumber,
			&client.CreatedAt,
			&client.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Get booking IDs for this client
		bookingIDs, err := r.getBookingIDsForClient(client.ID)
		if err != nil {
			return nil, err
		}
		client.BookingIDs = bookingIDs

		clients = append(clients, &client)
	}

	return clients, nil
}

func (r *ClientRepository) Update(client *domain.Client) error {
	query := `
		UPDATE clients
		SET name = ?, whatsapp_number = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.Exec(
		query,
		client.Name,
		client.WhatsAppNumber,
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

func (r *ClientRepository) getBookingIDsForClient(clientID domain.ClientID) ([]domain.BookingID, error) {
	query := `
		SELECT id
		FROM bookings
		WHERE client_id = ?
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookingIDs []domain.BookingID
	for rows.Next() {
		var bookingID domain.BookingID
		err := rows.Scan(&bookingID)
		if err != nil {
			return nil, err
		}
		bookingIDs = append(bookingIDs, bookingID)
	}

	return bookingIDs, nil
}
