package get_client

import (
	"errors"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/ports"
)

var (
	ErrClientNotFound = errors.New("client not found")
)

type Usecase struct {
	clientRepo ports.ClientRepository
}

func NewUsecase(clientRepo ports.ClientRepository) *Usecase {
	return &Usecase{
		clientRepo: clientRepo,
	}
}

func (u *Usecase) Execute(id domain.ClientID) (*domain.Client, error) {
	client, err := u.clientRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if client == nil {
		return nil, ErrClientNotFound
	}

	return client, nil
}
