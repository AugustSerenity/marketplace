package handler

import (
	"context"
	"net/url"

	"github.com/AugustSerenity/marketplace/internal/handler/model/ad"
	"github.com/AugustSerenity/marketplace/internal/handler/model/auth"
	"github.com/AugustSerenity/marketplace/internal/model"
)

type Service interface {
	RegisterUser(ctx context.Context, req *auth.RegistrationRequest) (*auth.RegistrationResponse, error)
	LoginUser(ctx context.Context, login, password string) (string, error)
	CreateAd(ctx context.Context, req ad.CreateRequest, userID int64) (*model.Ad, error)
	GetAds(ctx context.Context, req *ad.ListRequest, userID int64) ([]*model.AdWithAuthor, error)
	ParseListRequest(q url.Values) (ad.ListRequest, error)
}
