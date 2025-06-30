package get_all_clients

import (
	"github.com/mishkahtherapy/brain/core/domain"
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

func (u *Usecase) Execute() ([]*domain.Client, error) {
	return u.clientRepo.List()
}
