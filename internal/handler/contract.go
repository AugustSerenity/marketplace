package handler

import (
	"context"

	"github.com/AugustSerenity/marketplace/internal/model"
)

type Service interface {
	RegistrationUser(ctx context.Context, user *model.User) error
}
