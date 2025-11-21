package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/NailUsmanov/online_subscription/internal/models"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(dsn string) (*Storage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect db: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	// Настройка миграций
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate driver: %w", err)
	}

	// Инициализация мигратора
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise migrate driver: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

// Create - создает в базе новую строку о подписке.
func (s *Storage) Create(ctx context.Context, sub *models.Subscription) (int64, error) {
	err := s.db.QueryRowContext(ctx, InsertSubQuery,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
	).Scan(&sub.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to create new row: %w", err)
	}
	return sub.ID, nil
}

// Get - выдает данные по подписке по айди.
func (s *Storage) Get(ctx context.Context, id int64) (*models.Subscription, error) {
	var sub models.Subscription
	err := s.db.QueryRowContext(ctx, SelectSubQuery, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return &sub, nil
}

// Update - обновляет данные о подписке.
func (s *Storage) Update(ctx context.Context, sub *models.Subscription) error {
	res, err := s.db.ExecContext(ctx, UpdateQuery,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
		sub.ID)
	if err != nil {
		return fmt.Errorf("failed to update table: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Delete - удаляет сведения о подписке по айди.
func (s *Storage) Delete(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, DeleteQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed check affected rows: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// List - выдает все имеющиеся подписки пользователя.
func (s *Storage) List(ctx context.Context, userID string) ([]models.Subscription, error) {
	rows, err := s.db.QueryContext(ctx, ListQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get query from db: %w", err)
	}
	defer rows.Close()

	result := make([]models.Subscription, 0)

	for rows.Next() {
		res := models.Subscription{}
		err := rows.Scan(
			&res.ID,
			&res.ServiceName,
			&res.Price,
			&res.UserID,
			&res.StartDate,
			&res.EndDate,
		)
		if err != nil {
			return nil, fmt.Errorf("scan rows: %w", err)
		}
		result = append(result, res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return result, nil
}

// Sum - подсчитывает суммарную стоимость всех подписох за выбранный период.
func (s *Storage) Sum(ctx context.Context, filter models.SumSubscription) (int, error) {
	rows, err := s.db.QueryContext(ctx, SumQuery,
		filter.UserID,
		filter.ServiceName,
		filter.From,
		filter.To,
	)
	if err != nil {
		return 0, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	total := 0

	for rows.Next() {
		var price int
		var startDate time.Time
		var endDate *time.Time

		if err := rows.Scan(&price, &startDate, &endDate); err != nil {
			return 0, fmt.Errorf("scan rows: %w", err)
		}

		// считываю интервал пересечения
		from := filter.From
		to := filter.To

		// начало фактического периода - это максимум из start_date и from
		if startDate.After(from) {
			from = startDate
		}

		// конец фактического периода - минимум из end_date и to
		if endDate != nil && endDate.Before(to) {
			to = *endDate
		}

		// если после этого from > to, то пересечения нет
		if from.After(to) {
			continue
		}

		// считаю количество месяцев включительно
		y1, m1, _ := from.Date()
		y2, m2, _ := to.Date()
		months := (y2-y1)*12 + int(m2-m1) + 1
		if months < 0 {
			continue
		}
		total += price * months
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("rows iteration error: %w", err)
	}
	return total, nil
}
