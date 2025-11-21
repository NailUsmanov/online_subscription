package storage

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/NailUsmanov/online_subscription/internal/models"
)

func newTestStorage(t *testing.T) (*Storage, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}

	st := &Storage{db: db}

	cleanup := func() {
		_ = db.Close()
	}

	return st, mock, cleanup
}

func TestStorage_Create_OK(t *testing.T) {
	st, mock, cleanup := newTestStorage(t)
	defer cleanup()

	sub := &models.Subscription{
		ServiceName: "Netflix",
		Price:       999,
		UserID:      "user-1",
		StartDate:   time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
		// EndDate оставим nil
	}

	mock.ExpectQuery(regexp.QuoteMeta(InsertSubQuery)).
		WithArgs(sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(10)))

	id, err := st.Create(context.Background(), sub)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if id != 10 {
		t.Fatalf("id = %d, want 10", id)
	}
	if sub.ID != 10 {
		t.Fatalf("sub.ID = %d, want 10", sub.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestStorage_Create_DBError(t *testing.T) {
	st, mock, cleanup := newTestStorage(t)
	defer cleanup()

	sub := &models.Subscription{
		ServiceName: "Netflix",
		Price:       999,
		UserID:      "user-1",
		StartDate:   time.Now(),
	}

	mock.ExpectQuery(regexp.QuoteMeta(InsertSubQuery)).
		WithArgs(sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate).
		WillReturnError(errors.New("db error"))

	_, err := st.Create(context.Background(), sub)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestStorage_Get_OK(t *testing.T) {
	st, mock, cleanup := newTestStorage(t)
	defer cleanup()

	id := int64(5)
	start := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"id", "service_name", "price", "user_id", "start_date", "end_date",
	}).AddRow(id, "Netflix", 500, "user-1", start, end)

	mock.ExpectQuery(regexp.QuoteMeta(SelectSubQuery)).
		WithArgs(id).
		WillReturnRows(rows)

	sub, err := st.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if sub.ID != id {
		t.Fatalf("sub.ID = %d, want %d", sub.ID, id)
	}
	if sub.ServiceName != "Netflix" {
		t.Fatalf("service_name = %q, want %q", sub.ServiceName, "Netflix")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestStorage_Get_NoRows(t *testing.T) {
	st, mock, cleanup := newTestStorage(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(SelectSubQuery)).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrNoRows)

	_, err := st.Get(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err = %v, want sql.ErrNoRows", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestStorage_Update_OK(t *testing.T) {
	st, mock, cleanup := newTestStorage(t)
	defer cleanup()

	sub := &models.Subscription{
		ID:          7,
		ServiceName: "YouTube",
		Price:       200,
		UserID:      "user-1",
		StartDate:   time.Now(),
	}

	mock.ExpectExec(regexp.QuoteMeta(UpdateQuery)).
		WithArgs(sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate, sub.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := st.Update(context.Background(), sub)
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestStorage_Update_NotFound(t *testing.T) {
	st, mock, cleanup := newTestStorage(t)
	defer cleanup()

	sub := &models.Subscription{
		ID:          7,
		ServiceName: "YouTube",
		Price:       200,
		UserID:      "user-1",
		StartDate:   time.Now(),
	}

	mock.ExpectExec(regexp.QuoteMeta(UpdateQuery)).
		WithArgs(sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate, sub.ID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := st.Update(context.Background(), sub)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err = %v, want sql.ErrNoRows", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestStorage_Delete_OK(t *testing.T) {
	st, mock, cleanup := newTestStorage(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta(DeleteQuery)).
		WithArgs(int64(3)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := st.Delete(context.Background(), 3)
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestStorage_Delete_NotFound(t *testing.T) {
	st, mock, cleanup := newTestStorage(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta(DeleteQuery)).
		WithArgs(int64(3)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := st.Delete(context.Background(), 3)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err = %v, want sql.ErrNoRows", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestStorage_List_OK(t *testing.T) {
	st, mock, cleanup := newTestStorage(t)
	defer cleanup()

	userID := "user-1"

	rows := sqlmock.NewRows([]string{
		"id", "service_name", "price", "user_id", "start_date", "end_date",
	}).AddRow(int64(1), "Netflix", 100, userID, time.Now(), nil).
		AddRow(int64(2), "YouTube", 200, userID, time.Now(), nil)

	mock.ExpectQuery(regexp.QuoteMeta(ListQuery)).
		WithArgs(userID).
		WillReturnRows(rows)

	list, err := st.List(context.Background(), userID)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("len(list) = %d, want 2", len(list))
	}

	if list[0].ServiceName != "Netflix" || list[1].ServiceName != "YouTube" {
		t.Fatalf("unexpected services: %+v", list)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestStorage_List_RowError(t *testing.T) {
	st, mock, cleanup := newTestStorage(t)
	defer cleanup()

	userID := "user-1"

	rows := sqlmock.NewRows([]string{
		"id", "service_name", "price", "user_id", "start_date", "end_date",
	}).AddRow(int64(1), "Netflix", 100, userID, time.Now(), nil).
		RowError(0, errors.New("scan error")) // ошибка на первой строке

	mock.ExpectQuery(regexp.QuoteMeta(ListQuery)).
		WithArgs(userID).
		WillReturnRows(rows)

	_, err := st.List(context.Background(), userID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestStorage_Sum_OK(t *testing.T) {
	st, mock, cleanup := newTestStorage(t)
	defer cleanup()

	from := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC)

	filter := models.SumSubscription{
		UserID:      "user-1",
		ServiceName: "Netflix",
		From:        from,
		To:          to,
	}

	// первая подписка: с января по март, price=100 => 3 месяца => 300
	// вторая: с февраля по февраль, price=50 => 1 месяц => 50
	rows := sqlmock.NewRows([]string{"price", "start_date", "end_date"}).
		AddRow(100, from, nil).
		AddRow(50, time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC))

	mock.ExpectQuery(regexp.QuoteMeta(SumQuery)).
		WithArgs(filter.UserID, filter.ServiceName, filter.From, filter.To).
		WillReturnRows(rows)

	total, err := st.Sum(context.Background(), filter)
	if err != nil {
		t.Fatalf("Sum returned error: %v", err)
	}

	if total != 350 {
		t.Fatalf("total = %d, want 350", total)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestStorage_Sum_RowScanError(t *testing.T) {
	st, mock, cleanup := newTestStorage(t)
	defer cleanup()

	filter := models.SumSubscription{
		UserID:      "user-1",
		ServiceName: "Netflix",
		From:        time.Now(),
		To:          time.Now(),
	}

	rows := sqlmock.NewRows([]string{"price", "start_date", "end_date"}).
		AddRow("bad", time.Now(), nil) // тип не совпадает с int

	mock.ExpectQuery(regexp.QuoteMeta(SumQuery)).
		WithArgs(filter.UserID, filter.ServiceName, filter.From, filter.To).
		WillReturnRows(rows)

	_, err := st.Sum(context.Background(), filter)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
