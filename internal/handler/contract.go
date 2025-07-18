package handler

import (
	"context"

	"github.com/AugustSerenity/marketplace/internal/handler/model/ad"
	"github.com/AugustSerenity/marketplace/internal/model"
)

type Service interface {
	RegistrationUser(ctx context.Context, user *model.User) error
	LoginUser(ctx context.Context, login, password string) (string, error)
	CreateAd(ctx context.Context, ad *model.Ad) error
	GetAds(ctx context.Context, req *ad.ListRequest, userID int64) ([]*model.AdWithAuthor, error)
}
