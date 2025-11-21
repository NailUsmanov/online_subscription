package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	_ "github.com/NailUsmanov/online_subscription/internal/models"
	"github.com/NailUsmanov/online_subscription/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// CreateSubscriptionHandler - функция обработчик для создания подписки.
// CreateSubscriptionHandler godoc
// @Summary      Create subscription
// @Description  Создает новую подписку
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        input  body      service.CreateSubscription  true  "Subscription info"
// @Success      201    {object}  models.Subscription
// @Failure      400    {string}  string  "invalid input"
// @Failure      500    {string}  string  "internal server error"
// @Router       /subscriptions [post]
func CreateSubscriptionHandler(s service.Service, sugar *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Паршу JSON из запроса в структуру DTO для сервис слоя.
		var req service.CreateSubscription
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sugar.Errorf("cannot decode request JSON body: %v", err)
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}

		// Вызываю сервис слой и мапю ошибки.
		resp, err := s.Create(r.Context(), req)
		if err != nil {
			switch {

			case errors.Is(err, service.ErrInvalidServiceName):
				http.Error(w, "Invalid service name", http.StatusBadRequest)
			case errors.Is(err, service.ErrInvalidPrice):
				http.Error(w, "invalid price", http.StatusBadRequest)
			case errors.Is(err, service.ErrInvalidUserID):
				http.Error(w, "invalid userID", http.StatusBadRequest)
			case errors.Is(err, service.ErrInvalidStartDate):
				http.Error(w, "invalid start date", http.StatusBadRequest)
			case errors.Is(err, service.ErrInvalidEndDate):
				http.Error(w, "invalid end date", http.StatusBadRequest)
			default:
				sugar.Errorf("internal error in Create: %v", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}
		// Возвращаю ответ
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			sugar.Errorf("error encoding response: %v", err)
		}
		sugar.Infow("subscription created",
			"id", resp.ID,
			"user_id", resp.UserID,
			"service_name", resp.ServiceName,
		)
	}
}

// GetSubscriptionHandler - обработчик для получения данных о подписке.
// GetSubscriptionHandler godoc
// @Summary      Get subscription
// @Description  Возвращает подписку по ID
// @Tags         subscriptions
// @Produce      json
// @Param        id     path      int   true   "Subscription ID"
// @Success      200    {object}  models.Subscription
// @Failure      400    {string}  string  "invalid input"
// @Failure		 404    {string}  string  "not found"
// @Failure      500    {string}  string  "internal server error"
// @Router       /subscriptions/{id} [get]
func GetSubscriptionHandler(s service.Service, sugar *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Достать id из УРЛ: /subscriptions/{id}
		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			sugar.Errorf("invalid request id")
			http.Error(w, "invalid request id", http.StatusBadRequest)
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "id must be integer", http.StatusBadRequest)
			return
		}

		// Вызов сервиса.
		sub, err := s.Get(r.Context(), id)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidID):
				http.Error(w, "invalid id", http.StatusBadRequest)
			case errors.Is(err, service.ErrSubscriptionNotFound):
				http.Error(w, "subscription not found", http.StatusNotFound)
			default:
				sugar.Errorf("failed to get subscription id=%d: %v", id, err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Отправка ответа
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(sub); err != nil {
			sugar.Errorf("error encoding response: %v", err)
		}
	}
}

// UpdateSubscriptionHadnler - обработчик для обновления данных о подписке.
// UpdateSubscriptionHandler godoc
// @Summary      Update subscription
// @Description  Обновляет данные подписки по ID
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param		 id		path	  int	true "Subscription ID"
// @Param        input  body 	  service.CreateSubscription  true "Updated subscriptions"
// @Success      204    {string}  string  "no content"
// @Failure      400    {string}  string  "invalid input"
// @Failure		 404    {string}  string  "not found"
// @Failure      500    {string}  string  "internal server error"
// @Router       /subscriptions/{id} [put]
func UpdateSubscriptionHandler(s service.Service, sugar *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем id карточки подписки для обновления данных.
		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			sugar.Error("invalid request id")
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "id must be integer", http.StatusBadRequest)
			return
		}

		// Парсим данные из запроса
		var req service.CreateSubscription
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sugar.Errorf("cannot decode request JSON body: %v", err)
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}

		// Вызов сервиса.
		if err := s.Update(r.Context(), id, req); err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidID):
				http.Error(w, "invalid id", http.StatusBadRequest)
			case errors.Is(err, service.ErrInvalidServiceName):
				http.Error(w, "invalid service name", http.StatusBadRequest)
			case errors.Is(err, service.ErrInvalidPrice):
				http.Error(w, "invalid price", http.StatusBadRequest)
			case errors.Is(err, service.ErrInvalidUserID):
				http.Error(w, "invalid user id", http.StatusBadRequest)
			case errors.Is(err, service.ErrInvalidStartDate):
				http.Error(w, "invalid start date", http.StatusBadRequest)
			case errors.Is(err, service.ErrInvalidEndDate):
				http.Error(w, "invalid end date", http.StatusBadRequest)
			case errors.Is(err, service.ErrSubscriptionNotFound):
				http.Error(w, "subscription not found", http.StatusNotFound)
			default:
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			sugar.Errorf("update subscription failed: %v", err)
			return
		}

		// Отправка ответа
		w.WriteHeader(http.StatusNoContent)
		sugar.Infow("subscription updated",
			"id", id,
		)
	}
}

// DeleteSubscriptionsHandler - обработчик для удаления данных о подписке.
// DeleteSubscriptionsHandler godoc
// @Summary      Delete subscription
// @Description  Удаляет имеющуюся подписку по ID
// @Tags         subscriptions
// @Param        id  	path      int   true "Subscription ID"
// @Success      204    {string}  string  "no content"
// @Failure      400    {string}  string  "invalid input"
// @Failure		 404    {string}  string  "subscription not found"
// @Failure      500    {string}  string  "internal server error"
// @Router       /subscriptions/{id} [delete]
func DeleteSubscriptionsHandler(s service.Service, sugar *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			sugar.Error("invalid request id")
			http.Error(w, "invalid request id", http.StatusBadRequest)
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "id must be integer", http.StatusBadRequest)
			return
		}

		err = s.Delete(r.Context(), id)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidID):
				http.Error(w, "invalid id", http.StatusBadRequest)
			case errors.Is(err, service.ErrSubscriptionNotFound):
				http.Error(w, "subscription not found", http.StatusNotFound)
			default:
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			sugar.Errorf("failed to delete subscription id=%d: %v", id, err)
			return
		}

		// Отправка ответа
		w.WriteHeader(http.StatusNoContent)
		sugar.Infow("subscription deleted",
			"id", id,
		)
	}
}

// ListSubscriptionHandler - функция обработчик для выдачи всех подписок пользователя.
// ListSubscriptionHandler godoc
// @Summary      List subscriptions
// @Description  Возвращает все подписки пользователя
// @Tags         subscriptions
// @Produce      json
// @Param        user_id  query     string  true  "User ID"
// @Success      200      {array}   models.Subscription
// @Failure      400      {string}  string  "invalid input"
// @Failure      500      {string}  string  "internal server error"
// @Router       /subscriptions [get]
func ListSubscriptionHandler(s service.Service, sugar *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			sugar.Error("invalid userID in request")
			http.Error(w, "invalid userID in request", http.StatusBadRequest)
			return
		}

		subs, err := s.List(r.Context(), userID)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidUserID):
				http.Error(w, "invalid userID", http.StatusBadRequest)
			default:
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			sugar.Errorf("failed list subscriptions for user_id=%s: %v", userID, err)
			return
		}

		// Возвращаю ответ
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(subs); err != nil {
			sugar.Errorf("failed to encode list response: %v", err)
		}

		sugar.Infow("subscriptions listed",
			"user_id", userID,
			"count", len(subs),
		)
	}
}

// SumSubscriptionHandler - обработчик для выдачи суммы подписок пользователя с фильтрами.
// SumSubscriptionHandler godoc
// @Summary      Sum subscription cost
// @Description  Возвращает суммарную стоимость подписок пользователя с фильтрацией по сервису и периоду
// @Tags         subscriptions
// @Produce      json
// @Param        user_id       query     string  true   "User ID"
// @Param        service_name  query     string  true   "Service name"
// @Param        from  query     string  true   "Start date (MM-YYYY)"
// @Param        to    query     string  true   "End date (MM-YYYY)"
// @Success      200           {object}  map[string]int  "total sum"
// @Failure      400           {string}  string  "invalid input"
// @Failure      500           {string}  string  "internal server error"
// @Router       /subscriptions/sum [get]
func SumSubscriptionHandler(s service.Service, sugar *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаю данные из запроса
		q := r.URL.Query()

		userID := q.Get("user_id")
		serviceName := q.Get("service_name")
		from := q.Get("from")
		toStr := q.Get("to")

		if userID == "" || serviceName == "" || from == "" || toStr == "" {
			http.Error(w, "user_id, service_name, from, to are required", http.StatusBadRequest)
			return
		}

		filter := service.FilterForSumSubscription{
			UserID:      userID,
			ServiceName: serviceName,
			From:        from,
			To:          toStr,
		}

		// вызываю сервис
		total, err := s.Sum(r.Context(), filter)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidServiceName):
				http.Error(w, "invalid service name", http.StatusBadRequest)
			case errors.Is(err, service.ErrInvalidUserID):
				http.Error(w, "invalid userID", http.StatusBadRequest)
			case errors.Is(err, service.ErrInvalidStartDate):
				http.Error(w, "invalid start date", http.StatusBadRequest)
			case errors.Is(err, service.ErrInvalidEndDate):
				http.Error(w, "invalid end date", http.StatusBadRequest)
			default:
				sugar.Errorf("failed sum function: %v", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Возвращаю ответ
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		resp := map[string]int{"total": total}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			sugar.Errorf("failed to encode sum response: %v", err)
		}

		sugar.Infow("subscriptions sum calculated",
			"user_id", userID,
			"service_name", serviceName,
			"from", from,
			"to", toStr,
			"total", total,
		)
	}
}
