package ports

import "github.com/mishkahtherapy/brain/core/domain"

type ClientRepository interface {
	GetByID(id string) (*domain.Client, error)
	Create(client *domain.Client) error
	Update(client *domain.Client) error
	Delete(id string) error
	List() ([]*domain.Client, error)
}
