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

func (u *Usecase) Execute(id domain.ClientID) (*client.Client, error) {
	client, err := u.clientRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if client == nil {
		return nil, common.ErrClientNotFound
	}

	return client, nil
}
