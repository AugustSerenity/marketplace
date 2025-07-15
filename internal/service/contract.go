package service

import (
	"context"

	"github.com/AugustSerenity/marketplace/internal/model"
)

type Storage interface {
	CreateUser(ctx context.Context, user *model.User) error
}
