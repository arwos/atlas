package dhcp

import (
	"context"

	"go.arwos.org/atlas/pkg/database"
)

type Service struct {
	repo *repository
}

func NewService(db *database.Service) *Service {
	return &Service{
		repo: &repository{
			db: db,
		},
	}
}

func (s *Service) Up(ctx context.Context) error {
	// Implement the logic to start the DHCP service here
	return nil
}

func (s *Service) Down() error {
	// Implement the logic to stop the DHCP service here
	return nil
}
