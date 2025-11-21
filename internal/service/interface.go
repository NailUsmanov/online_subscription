package service

import (
	"context"

	"github.com/NailUsmanov/online_subscription/internal/models"
)

// SubscriptionRepository - описывает поведение данных в БД.
type SubscriptionRepository interface {
	Create(ctx context.Context, sub *models.Subscription) (int64, error)
	Get(ctx context.Context, id int64) (sub *models.Subscription, err error)
	Update(ctx context.Context, sub *models.Subscription) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, userID string) ([]models.Subscription, error)
	Sum(ctx context.Context, filter models.SumSubscription) (int, error)
}
