package service

import (
	"context"

	"github.com/AugustSerenity/marketplace/internal/model"
)

type Service struct {
	storage Storage
}

func New(st Storage) *Service {
	return &Service{
		storage: st,
	}
}

func (s *Service) RegistrationUser(ctx context.Context, user *model.User) error {
	return s.storage.CreateUser(ctx, user)
}
