package get_all_clients

import (
	"slices"

	"github.com/mishkahtherapy/brain/core/domain"
	"github.com/mishkahtherapy/brain/core/domain/client"
	"github.com/mishkahtherapy/brain/core/ports"
)

type Input struct {
	WhatsApp domain.WhatsAppNumber
	Ids      []domain.ClientID
}

type Usecase struct {
	clientRepo ports.ClientRepository
}

func NewUsecase(clientRepo ports.ClientRepository) *Usecase {
	return &Usecase{
		clientRepo: clientRepo,
	}
}

func (u *Usecase) Execute(input Input) ([]*client.Client, error) {
	clients, err := u.clientRepo.List()
	if err != nil {
		return nil, err
	}

	filteredClients := make([]*client.Client, 0)
	for _, client := range clients {
		if input.WhatsApp != "" && client.WhatsAppNumber != input.WhatsApp {
			continue
		}
		if len(input.Ids) > 0 && !slices.Contains(input.Ids, client.ID) {
			continue
		}
		filteredClients = append(filteredClients, client)
	}

	return filteredClients, nil
}
