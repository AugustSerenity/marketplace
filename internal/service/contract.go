package service

import (
	"context"

	"github.com/AugustSerenity/marketplace/internal/model"
)

type Storage interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)
	CreateAd(ctx context.Context, ad *model.Ad) error
}
