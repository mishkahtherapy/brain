package get_all_clients

import (
	"github.com/mishkahtherapy/brain/core/domain/client"
	"github.com/mishkahtherapy/brain/core/ports"
)

type Usecase struct {
	clientRepo ports.ClientRepository
}

func NewUsecase(clientRepo ports.ClientRepository) *Usecase {
	return &Usecase{
		clientRepo: clientRepo,
	}
}

func (u *Usecase) Execute() ([]*client.Client, error) {
	return u.clientRepo.List()
}
