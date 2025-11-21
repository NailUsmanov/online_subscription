package models

import "time"

// Subscription - описывает сущность подписка.
type Subscription struct {
	ID          int64      `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      string     `json:"user_id"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}

// SumSubscription - используется для задачи фильтрации и показания общей суммы подписок за выбранный период.
type SumSubscription struct {
	UserID      string    `json:"user_id"`
	ServiceName string    `json:"service_name"`
	From        time.Time `json:"from"`
	To          time.Time `json:"to"`
}
