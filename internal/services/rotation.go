package services

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/kristina71/otus_project/internal/stats"
	"github.com/kristina71/otus_project/internal/storage"
)

//nolint:lll
//go:generate mockgen --build_flags=--mod=mod -destination=./mock_types_test.go -package=services_test . Repository,EventsPublisher
type Repository interface {
	AddSlot(ctx context.Context, description string) (string, error)
	GetSlotByID(ctx context.Context, id string) (storage.Slot, error)
	DeleteSlot(ctx context.Context, id string) error
	AddBanner(ctx context.Context, description string) (string, error)
	GetBannerByID(ctx context.Context, id string) (storage.Banner, error)
	DeleteBanner(ctx context.Context, id string) error
	AddBannerToSlot(ctx context.Context, slotID, bannerID string) error
	DeleteBannerFromSlot(ctx context.Context, slotID, bannerID string) error
	AddGroup(ctx context.Context, description string) (string, error)
	GetGroupByID(ctx context.Context, groupID string) (storage.SocialGroup, error)
	DeleteGroup(ctx context.Context, id string) error
	PersistClick(ctx context.Context, slotID, groupID, bannerID string) error
	PersistShow(ctx context.Context, slotID, groupID, bannerID string) error
	FindSlotBannerStats(ctx context.Context, slotID, groupID string) ([]storage.SlotBannerStat, error)
}

type EventsPublisher interface {
	Publish(msg stats.Message) error
}

type RotationService struct {
	repo      Repository
	publisher EventsPublisher
}

func NewRotationService(repo Repository, publisher EventsPublisher) RotationService {
	return RotationService{repo, publisher}
}

func (r RotationService) AddSlot(ctx context.Context, description string) (storage.Slot, error) {
	slotID, err := r.repo.AddSlot(ctx, description)
	if err != nil {
		return storage.Slot{}, fmt.Errorf("error during slot creation: %w", err)
	}
	return storage.Slot{
		ID:          slotID,
		Description: description,
	}, nil
}

func (r RotationService) DeleteSlot(ctx context.Context, slotID string) error {
	if err := r.repo.DeleteSlot(ctx, slotID); err != nil {
		return fmt.Errorf("error during slot deleting: %w", err)
	}
	return nil
}

func (r RotationService) AddBannerToSlot(ctx context.Context, slotID, bannerID string) error {
	err := r.repo.AddBannerToSlot(ctx, slotID, bannerID)
	if err != nil {
		return fmt.Errorf("error during adding banner to slot: %w", err)
	}
	return nil
}

func (r RotationService) DeleteBannerFromSlot(ctx context.Context, bannerID, slotID string) error {
	if err := r.repo.DeleteBannerFromSlot(ctx, bannerID, slotID); err != nil {
		return fmt.Errorf("errod during deleting banner from slot: %w", err)
	}
	return nil
}

func (r RotationService) AddBanner(ctx context.Context, description string) (storage.Banner, error) {
	bannerID, err := r.repo.AddBanner(ctx, description)
	if err != nil {
		return storage.Banner{}, fmt.Errorf("error during creating banner: %w", err)
	}
	return storage.Banner{
		ID:          bannerID,
		Description: description,
	}, nil
}

func (r RotationService) DeleteBanner(ctx context.Context, bannerID string) error {
	if err := r.repo.DeleteBanner(ctx, bannerID); err != nil {
		return fmt.Errorf("error during deleting banner: %w", err)
	}
	return nil
}

func (r RotationService) AddGroup(ctx context.Context, description string) (storage.SocialGroup, error) {
	groupID, err := r.repo.AddGroup(ctx, description)
	if err != nil {
		return storage.SocialGroup{}, fmt.Errorf("error during adding group: %w", err)
	}
	return storage.SocialGroup{
		ID:          groupID,
		Description: description,
	}, nil
}

func (r RotationService) DeleteGroup(ctx context.Context, groupID string) error {
	if err := r.repo.DeleteGroup(ctx, groupID); err != nil {
		return fmt.Errorf("error during deleting group by id: %w", err)
	}
	return nil
}

func (r RotationService) PersistClick(ctx context.Context, slotID, groupID, bannerID string) error {
	if err := r.repo.PersistClick(ctx, slotID, groupID, bannerID); err != nil {
		return fmt.Errorf("failed to persist banner click stats: %w", err)
	}
	if err := r.publisher.Publish(stats.Message{
		BannerID:  bannerID,
		SlotID:    slotID,
		GroupID:   groupID,
		Type:      "click",
		Timestamp: time.Now(),
	}); err != nil {
		return fmt.Errorf("failed to publish click event stats to rabbit queue: %w", err)
	}
	return nil
}

func (r RotationService) NextBannerID(ctx context.Context, slotID, groupID string) (string, error) {
	bannerStats, err := r.repo.FindSlotBannerStats(ctx, slotID, groupID)
	if err != nil {
		return "", fmt.Errorf("failed to get banner statistics for a slot: %w", err)
	}
	if len(bannerStats) == 0 {
		return "", storage.ErrNoBannersFoundForSlot
	}

	maxBannerID := calculateNextBannerID(bannerStats)
	if err := r.repo.PersistShow(ctx, slotID, groupID, maxBannerID); err != nil {
		return "", fmt.Errorf("failed to store banner show: %w", err)
	}
	if err := r.publisher.Publish(stats.Message{
		BannerID:  maxBannerID,
		SlotID:    slotID,
		GroupID:   groupID,
		Type:      "show",
		Timestamp: time.Now(),
	}); err != nil {
		return "", fmt.Errorf("failed to publish click event stats to rabbit queue: %w", err)
	}

	return maxBannerID, nil
}

func calculateNextBannerID(bannerStats []storage.SlotBannerStat) string {
	// all available banners should be shown at least once
	for _, bannerStat := range bannerStats {
		if bannerStat.GetShows() == 0 {
			return bannerStat.BannerID
		}
	}

	// show banner which has max targetFunction value
	totalBannerShows := countTotalShowsAmount(bannerStats)
	maxTargetValue := 0.0
	maxBannerID := bannerStats[0].BannerID
	for _, bannerStat := range bannerStats {
		bannerClicks, bannerShows := bannerStat.GetClicks(), bannerStat.GetShows()
		targetValue := targetFunction(float64(bannerClicks), float64(bannerShows), float64(totalBannerShows))
		if big.NewFloat(targetValue).Cmp(big.NewFloat(maxTargetValue)) > 0 {
			maxTargetValue = targetValue
			maxBannerID = bannerStat.BannerID
		}
	}
	return maxBannerID
}

func countTotalShowsAmount(stats []storage.SlotBannerStat) int64 {
	var totalShows int64
	for _, stat := range stats {
		totalShows += stat.GetShows()
	}
	return totalShows
}

func targetFunction(clickCount, showCount, totalShowCount float64) float64 {
	avgBannerIncome := clickCount / showCount
	return avgBannerIncome + math.Sqrt((2.0*math.Log(totalShowCount))/showCount)
}
