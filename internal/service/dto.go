package service

// CreateSubscription - структура для входящих данных из хендлера.
type CreateSubscription struct {
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date"`
}

// FilterForSumSubscription - структура для фильтрации данных из хендлера для функции суммирования.
type FilterForSumSubscription struct {
	UserID      string
	ServiceName string
	From        string
	To          string
}
