package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/NailUsmanov/online_subscription/internal/models"
	"github.com/google/uuid"
)

var (
	ErrInvalidID            = errors.New("invalid id")
	ErrInvalidServiceName   = errors.New("invalid service_name")
	ErrInvalidPrice         = errors.New("invalid price")
	ErrInvalidUserID        = errors.New("invalid user_id")
	ErrInvalidStartDate     = errors.New("invalid start_date")
	ErrInvalidEndDate       = errors.New("invalid end_date")
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

// Service - структура сервис слоя для обработки информации из хендлера, валидации и связи с БД-слоем.
type Service struct {
	repo SubscriptionRepository
}

// NewSubscriptionService - конструктор сервиса.
func NewService(repo SubscriptionRepository) Service {
	return Service{repo: repo}
}

// Create - создает новую подписку с валидацией и преобразованием данных.
func (s *Service) Create(ctx context.Context, data CreateSubscription) (*models.Subscription, error) {
	// Требуется валидация всех параметров
	data.ServiceName = strings.TrimSpace(data.ServiceName)
	if data.ServiceName == "" {
		return nil, ErrInvalidServiceName
	}

	if data.Price <= 0 {
		return nil, ErrInvalidPrice
	}

	if _, err := uuid.Parse(data.UserID); err != nil {
		return nil, ErrInvalidUserID
	}

	start, err := time.Parse("01-2006", data.StartDate)
	if err != nil {
		return nil, ErrInvalidStartDate
	}

	var end *time.Time
	// проверяем наличие поля в присланном JSON и затем, что оно не пустое
	if data.EndDate != nil && *data.EndDate != "" {
		endParse, err := time.Parse("01-2006", *data.EndDate)
		if err != nil {
			return nil, ErrInvalidEndDate
		}

		if endParse.Before(start) {
			return nil, ErrInvalidEndDate
		}
		end = &endParse
	}

	// Собираю модель, для отправки в репо
	sub := &models.Subscription{
		ServiceName: data.ServiceName,
		Price:       data.Price,
		UserID:      data.UserID,
		StartDate:   start,
		EndDate:     end,
	}

	// Вызываю метод репо
	id, err := s.repo.Create(ctx, sub)
	if err != nil {
		return nil, err
	}
	sub.ID = id

	return sub, nil
}

// Get - получает подписку для выдачи пользователю.
func (s *Service) Get(ctx context.Context, id int64) (*models.Subscription, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}

	sub, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, err
	}

	return sub, nil
}

// Update - обновляет данные о подписке, с валидацией входных параметров.
func (s *Service) Update(ctx context.Context, id int64, sub CreateSubscription) error {
	// Требуется валидация всех параметров
	if id <= 0 {
		return ErrInvalidID
	}

	sub.ServiceName = strings.TrimSpace(sub.ServiceName)
	if sub.ServiceName == "" {
		return ErrInvalidServiceName
	}

	if sub.Price <= 0 {
		return ErrInvalidPrice
	}

	if _, err := uuid.Parse(sub.UserID); err != nil {
		return ErrInvalidUserID
	}

	start, err := time.Parse("01-2006", sub.StartDate)
	if err != nil {
		return ErrInvalidStartDate
	}

	var end *time.Time
	// проверяем наличие поля в присланном JSON и затем, что оно не пустое
	if sub.EndDate != nil && *sub.EndDate != "" {
		endParse, err := time.Parse("01-2006", *sub.EndDate)
		if err != nil {
			return ErrInvalidEndDate
		}

		if endParse.Before(start) {
			return ErrInvalidEndDate
		}
		end = &endParse
	}

	// Собираю модель, для отправки в репо
	subToRepo := &models.Subscription{
		ID:          id,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID,
		StartDate:   start,
		EndDate:     end,
	}

	// Вызываю репо
	err = s.repo.Update(ctx, subToRepo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSubscriptionNotFound
		}
		return err
	}
	return nil
}

// Delete - удаляет данные о подписке с валидацией ID.
func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidID
	}

	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSubscriptionNotFound
		}
		return err
	}
	return nil
}

// List - проверяет userID и выдает все подписки пользователя.
func (s *Service) List(ctx context.Context, userID string) ([]models.Subscription, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return nil, ErrInvalidUserID
	}
	result, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) Sum(ctx context.Context, filter FilterForSumSubscription) (int, error) {
	// Валидирую данные
	filter.ServiceName = strings.TrimSpace(filter.ServiceName)
	if filter.ServiceName == "" {
		return 0, ErrInvalidServiceName
	}

	if _, err := uuid.Parse(filter.UserID); err != nil {
		return 0, ErrInvalidUserID
	}

	from, err := time.Parse("01-2006", filter.From)
	if err != nil {
		return 0, ErrInvalidStartDate
	}

	to, err := time.Parse("01-2006", filter.To)
	if err != nil {
		return 0, ErrInvalidEndDate
	}

	if to.Before(from) {
		return 0, ErrInvalidEndDate
	}

	// Собираю модель для отправки в репо
	sumFilter := models.SumSubscription{
		UserID:      filter.UserID,
		ServiceName: filter.ServiceName,
		From:        from,
		To:          to,
	}

	totalSum, err := s.repo.Sum(ctx, sumFilter)
	if err != nil {
		return 0, err
	}
	return totalSum, nil
}
