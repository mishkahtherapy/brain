package ports

import (
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/client"
)

type ClientRepository interface {
	Create(client *client.Client) error
	GetByID(id domain.ClientID) (*client.Client, error)
	GetByWhatsAppNumber(whatsAppNumber domain.WhatsAppNumber) (*client.Client, error)
	List() ([]*client.Client, error)
	Update(client *client.Client) error
	Delete(id domain.ClientID) error
	UpdateTimezone(id domain.ClientID, timezone domain.Timezone) error
}
