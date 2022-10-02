package server

import (
	"context"

	"github.com/kristina71/otus_project/internal/storage"
)

type Application interface {
	AddSlot(ctx context.Context, description string) (storage.Slot, error)
	DeleteSlot(ctx context.Context, slotID string) error
	AddBannerToSlot(ctx context.Context, slotID, bannerID string) error
	DeleteBannerFromSlot(ctx context.Context, bannerID, slotID string) error
	AddBanner(ctx context.Context, description string) (storage.Banner, error)
	DeleteBanner(ctx context.Context, bannerID string) error
	AddGroup(ctx context.Context, description string) (storage.SocialGroup, error)
	DeleteGroup(ctx context.Context, groupID string) error
	PersistClick(ctx context.Context, slotID, groupID, bannerID string) error
	NextBannerID(ctx context.Context, slotID, groupID string) (string, error)
}
