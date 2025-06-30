package ports

import "github.com/mishkahtherapy/brain/core/domain"

type ClientRepository interface {
	Create(client *domain.Client) error
	GetByID(id domain.ClientID) (*domain.Client, error)
	GetByWhatsAppNumber(whatsAppNumber domain.WhatsAppNumber) (*domain.Client, error)
	List() ([]*domain.Client, error)
	Update(client *domain.Client) error
	Delete(id domain.ClientID) error
}
