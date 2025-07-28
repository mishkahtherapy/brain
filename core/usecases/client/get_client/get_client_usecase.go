package get_client

import (
	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/client"
	"github.com/mishkahtherapy/brain/core/ports"
	"github.com/mishkahtherapy/brain/core/usecases/common"
)

type Usecase struct {
	clientRepo ports.ClientRepository
}

func NewUsecase(clientRepo ports.ClientRepository) *Usecase {
	return &Usecase{
		clientRepo: clientRepo,
	}
}

func (u *Usecase) Execute(ids []domain.ClientID) ([]*client.Client, error) {
	clients, err := u.clientRepo.BulkGetByID(ids)
	if err != nil {
		return nil, err
	}

	if len(clients) == 0 {
		return nil, common.ErrClientNotFound
	}

	return clients, nil
}
