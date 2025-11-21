package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/NailUsmanov/online_subscription/internal/models"
)

// простой стаб для репозитория
type stubRepo struct {
	createSub *models.Subscription
	createID  int64
	createErr error

	getID  int64
	getSub *models.Subscription
	getErr error

	updateSub *models.Subscription
	updateErr error

	deleteID  int64
	deleteErr error

	listUserID string
	listSubs   []models.Subscription
	listErr    error

	sumFilter models.SumSubscription
	sumTotal  int
	sumErr    error
}

func (r *stubRepo) Create(_ context.Context, sub *models.Subscription) (int64, error) {
	r.createSub = sub
	return r.createID, r.createErr
}

func (r *stubRepo) Get(_ context.Context, id int64) (*models.Subscription, error) {
	r.getID = id
	return r.getSub, r.getErr
}

func (r *stubRepo) Update(_ context.Context, sub *models.Subscription) error {
	r.updateSub = sub
	return r.updateErr
}

func (r *stubRepo) Delete(_ context.Context, id int64) error {
	r.deleteID = id
	return r.deleteErr
}

func (r *stubRepo) List(_ context.Context, userID string) ([]models.Subscription, error) {
	r.listUserID = userID
	return r.listSubs, r.listErr
}

func (r *stubRepo) Sum(_ context.Context, f models.SumSubscription) (int, error) {
	r.sumFilter = f
	return r.sumTotal, r.sumErr
}

func TestService_Create_ValidationErrors(t *testing.T) {
	repo := &stubRepo{}
	s := NewService(repo)
	ctx := context.Background()

	tests := []struct {
		name string
		data CreateSubscription
		want error
	}{
		{
			name: "empty service name",
			data: CreateSubscription{
				ServiceName: "   ",
				Price:       100,
				UserID:      "550e8400-e29b-41d4-a716-446655440000",
				StartDate:   "01-2024",
			},
			want: ErrInvalidServiceName,
		},
		{
			name: "non positive price",
			data: CreateSubscription{
				ServiceName: "Netflix",
				Price:       0,
				UserID:      "550e8400-e29b-41d4-a716-446655440000",
				StartDate:   "01-2024",
			},
			want: ErrInvalidPrice,
		},
		{
			name: "bad user id",
			data: CreateSubscription{
				ServiceName: "Netflix",
				Price:       100,
				UserID:      "not-uuid",
				StartDate:   "01-2024",
			},
			want: ErrInvalidUserID,
		},
		{
			name: "bad start date",
			data: CreateSubscription{
				ServiceName: "Netflix",
				Price:       100,
				UserID:      "550e8400-e29b-41d4-a716-446655440000",
				StartDate:   "2024-01-01",
			},
			want: ErrInvalidStartDate,
		},
		{
			name: "bad end date format",
			data: func() CreateSubscription {
				d := "2024-01-01"
				return CreateSubscription{
					ServiceName: "Netflix",
					Price:       100,
					UserID:      "550e8400-e29b-41d4-a716-446655440000",
					StartDate:   "01-2024",
					EndDate:     &d,
				}
			}(),
			want: ErrInvalidEndDate,
		},
		{
			name: "end before start",
			data: func() CreateSubscription {
				d := "01-2023"
				return CreateSubscription{
					ServiceName: "Netflix",
					Price:       100,
					UserID:      "550e8400-e29b-41d4-a716-446655440000",
					StartDate:   "01-2024",
					EndDate:     &d,
				}
			}(),
			want: ErrInvalidEndDate,
		},
	}

	for _, tt := range tests {
		_, err := s.Create(ctx, tt.data)
		if !errors.Is(err, tt.want) {
			t.Fatalf("%s: err = %v, want %v", tt.name, err, tt.want)
		}
	}
}

func TestService_Create_Success(t *testing.T) {
	repo := &stubRepo{
		createID:  10,
		createErr: nil,
	}
	s := NewService(repo)
	ctx := context.Background()

	end := "02-2024"
	in := CreateSubscription{
		ServiceName: "  Netflix  ",
		Price:       999,
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		StartDate:   "01-2024",
		EndDate:     &end,
	}

	got, err := s.Create(ctx, in)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if got.ID != 10 {
		t.Fatalf("ID = %d, want 10", got.ID)
	}

	if repo.createSub == nil {
		t.Fatalf("Create did not call repo")
	}

	if repo.createSub.ServiceName != "Netflix" {
		t.Fatalf("ServiceName sent to repo = %q, want %q", repo.createSub.ServiceName, "Netflix")
	}
}

func TestService_Get_InvalidID(t *testing.T) {
	repo := &stubRepo{}
	s := NewService(repo)

	_, err := s.Get(context.Background(), 0)
	if !errors.Is(err, ErrInvalidID) {
		t.Fatalf("err = %v, want %v", err, ErrInvalidID)
	}

	if repo.getID != 0 {
		t.Fatalf("repo.Get should not be called on invalid id")
	}
}

func TestService_Get_NotFound(t *testing.T) {
	repo := &stubRepo{
		getErr: sql.ErrNoRows,
	}
	s := NewService(repo)

	_, err := s.Get(context.Background(), 1)
	if !errors.Is(err, ErrSubscriptionNotFound) {
		t.Fatalf("err = %v, want %v", err, ErrSubscriptionNotFound)
	}
}

func TestService_Get_Success(t *testing.T) {
	sub := &models.Subscription{ID: 5}
	repo := &stubRepo{
		getSub: sub,
	}
	s := NewService(repo)

	got, err := s.Get(context.Background(), 5)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if got.ID != 5 {
		t.Fatalf("ID = %d, want 5", got.ID)
	}

	if repo.getID != 5 {
		t.Fatalf("repo.Get called with id = %d, want 5", repo.getID)
	}
}

func TestService_Update_ValidationErrors(t *testing.T) {
	repo := &stubRepo{}
	s := NewService(repo)
	ctx := context.Background()

	ok := CreateSubscription{
		ServiceName: "Netflix",
		Price:       100,
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		StartDate:   "01-2024",
	}

	tests := []struct {
		name string
		id   int64
		data CreateSubscription
		want error
	}{
		{
			name: "invalid id",
			id:   0,
			data: ok,
			want: ErrInvalidID,
		},
		{
			name: "empty service name",
			id:   1,
			data: func() CreateSubscription {
				d := ok
				d.ServiceName = " "
				return d
			}(),
			want: ErrInvalidServiceName,
		},
		{
			name: "price <= 0",
			id:   1,
			data: func() CreateSubscription {
				d := ok
				d.Price = 0
				return d
			}(),
			want: ErrInvalidPrice,
		},
		{
			name: "bad user id",
			id:   1,
			data: func() CreateSubscription {
				d := ok
				d.UserID = "bad"
				return d
			}(),
			want: ErrInvalidUserID,
		},
		{
			name: "bad start date",
			id:   1,
			data: func() CreateSubscription {
				d := ok
				d.StartDate = "2024-01-01"
				return d
			}(),
			want: ErrInvalidStartDate,
		},
		{
			name: "bad end date",
			id:   1,
			data: func() CreateSubscription {
				d := ok
				e := "2024-01-01"
				d.EndDate = &e
				return d
			}(),
			want: ErrInvalidEndDate,
		},
		{
			name: "end before start",
			id:   1,
			data: func() CreateSubscription {
				d := ok
				e := "01-2023"
				d.EndDate = &e
				return d
			}(),
			want: ErrInvalidEndDate,
		},
	}

	for _, tt := range tests {
		err := s.Update(ctx, tt.id, tt.data)
		if !errors.Is(err, tt.want) {
			t.Fatalf("%s: err = %v, want %v", tt.name, err, tt.want)
		}
	}
}

func TestService_Update_NotFound(t *testing.T) {
	repo := &stubRepo{
		updateErr: sql.ErrNoRows,
	}
	s := NewService(repo)

	err := s.Update(context.Background(), 1, CreateSubscription{
		ServiceName: "Netflix",
		Price:       100,
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		StartDate:   "01-2024",
	})
	if !errors.Is(err, ErrSubscriptionNotFound) {
		t.Fatalf("err = %v, want %v", err, ErrSubscriptionNotFound)
	}
}

func TestService_Update_Success(t *testing.T) {
	repo := &stubRepo{}
	s := NewService(repo)

	err := s.Update(context.Background(), 3, CreateSubscription{
		ServiceName: "Netflix",
		Price:       100,
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		StartDate:   "01-2024",
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	if repo.updateSub == nil {
		t.Fatalf("repo.Update not called")
	}

	if repo.updateSub.ID != 3 {
		t.Fatalf("repo.Update got ID = %d, want 3", repo.updateSub.ID)
	}
}

func TestService_Delete_InvalidID(t *testing.T) {
	repo := &stubRepo{}
	s := NewService(repo)

	err := s.Delete(context.Background(), -1)
	if !errors.Is(err, ErrInvalidID) {
		t.Fatalf("err = %v, want %v", err, ErrInvalidID)
	}

	if repo.deleteID != 0 {
		t.Fatalf("repo.Delete should not be called on invalid id")
	}
}

func TestService_Delete_NotFound(t *testing.T) {
	repo := &stubRepo{
		deleteErr: sql.ErrNoRows,
	}
	s := NewService(repo)

	err := s.Delete(context.Background(), 1)
	if !errors.Is(err, ErrSubscriptionNotFound) {
		t.Fatalf("err = %v, want %v", err, ErrSubscriptionNotFound)
	}
}

func TestService_Delete_Success(t *testing.T) {
	repo := &stubRepo{}
	s := NewService(repo)

	err := s.Delete(context.Background(), 7)
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	if repo.deleteID != 7 {
		t.Fatalf("repo.Delete called with id = %d, want 7", repo.deleteID)
	}
}

func TestService_List_InvalidUserID(t *testing.T) {
	repo := &stubRepo{}
	s := NewService(repo)

	_, err := s.List(context.Background(), "bad-uuid")
	if !errors.Is(err, ErrInvalidUserID) {
		t.Fatalf("err = %v, want %v", err, ErrInvalidUserID)
	}
}

func TestService_List_Success(t *testing.T) {
	subs := []models.Subscription{{ID: 1}, {ID: 2}}
	repo := &stubRepo{
		listSubs: subs,
	}
	s := NewService(repo)

	userID := "550e8400-e29b-41d4-a716-446655440000"
	got, err := s.List(context.Background(), userID)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}

	if repo.listUserID != userID {
		t.Fatalf("repo.List called with %q, want %q", repo.listUserID, userID)
	}
}

func TestService_Sum_ValidationErrors(t *testing.T) {
	repo := &stubRepo{}
	s := NewService(repo)
	ctx := context.Background()

	base := FilterForSumSubscription{
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		ServiceName: "Netflix",
		From:        "01-2024",
		To:          "02-2024",
	}

	tests := []struct {
		name   string
		filter FilterForSumSubscription
		want   error
	}{
		{
			name: "empty service name",
			filter: func() FilterForSumSubscription {
				f := base
				f.ServiceName = " "
				return f
			}(),
			want: ErrInvalidServiceName,
		},
		{
			name: "bad user id",
			filter: func() FilterForSumSubscription {
				f := base
				f.UserID = "bad"
				return f
			}(),
			want: ErrInvalidUserID,
		},
		{
			name: "bad from date",
			filter: func() FilterForSumSubscription {
				f := base
				f.From = "2024-01-01"
				return f
			}(),
			want: ErrInvalidStartDate,
		},
		{
			name: "bad to date",
			filter: func() FilterForSumSubscription {
				f := base
				f.To = "2024-02-01"
				return f
			}(),
			want: ErrInvalidEndDate,
		},
		{
			name: "to before from",
			filter: func() FilterForSumSubscription {
				f := base
				f.To = "01-2023"
				return f
			}(),
			want: ErrInvalidEndDate,
		},
	}

	for _, tt := range tests {
		_, err := s.Sum(ctx, tt.filter)
		if !errors.Is(err, tt.want) {
			t.Fatalf("%s: err = %v, want %v", tt.name, err, tt.want)
		}
	}
}

func TestService_Sum_Success(t *testing.T) {
	repo := &stubRepo{
		sumTotal: 1500,
	}
	s := NewService(repo)

	filter := FilterForSumSubscription{
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		ServiceName: "Netflix",
		From:        "01-2024",
		To:          "02-2024",
	}

	total, err := s.Sum(context.Background(), filter)
	if err != nil {
		t.Fatalf("Sum returned error: %v", err)
	}

	if total != 1500 {
		t.Fatalf("total = %d, want 1500", total)
	}

	// минимальная проверка, что в репо ушли правильные даты и user_id
	if repo.sumFilter.UserID != filter.UserID {
		t.Fatalf("repo.Sum userID = %q, want %q", repo.sumFilter.UserID, filter.UserID)
	}

	if !repo.sumFilter.From.Equal(mustParse("01-2024")) {
		t.Fatalf("repo.sumFilter.From = %v, want 01-2024", repo.sumFilter.From)
	}
}

func mustParse(s string) time.Time {
	tm, err := time.Parse("01-2006", s)
	if err != nil {
		panic(err)
	}
	return tm
}
