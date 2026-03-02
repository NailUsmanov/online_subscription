package storage

// InsertSubQuery запрос для добавления в таблицу новой подписки.
var InsertSubQuery string = `
	INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id
`

// SelectSubQuery - запрос для получения информации о подписке по ID.
var SelectSubQuery string = `
	SELECT id, service_name, price, user_id, start_date, end_date
	FROM subscriptions
	WHERE id = $1;
`

// UpdateQuery - запрос обновления данных в таблице.
var UpdateQuery string = `
UPDATE subscriptions
SET service_name = $1, price = $2, user_id = $3, start_date = $4, end_date = $5
WHERE id = $6;
`

// DeleteQuery - запрос на удаление данных о подписке.
var DeleteQuery string = `
	DELETE FROM subscriptions
	WHERE id = $1;
`

// ListQuery - запрос на выдачу списка всех подписок конкретного пользователя.
var ListQuery string = `
    SELECT id, service_name, price, user_id, start_date, end_date
    FROM subscriptions
    WHERE user_id = $1
    ORDER BY id
    LIMIT $2 OFFSET $3
`

var CountQuery = `
		SELECT COUNT(*) FROM subscriptions WHERE user_id = $1
	`

// SumQuery - запрос для выдачи стоимости подписки пользователя за выбранный период.
var SumQuery = `
		WITH months_in_period AS (
			SELECT generate_series(
				date_trunc('month', $3::date),
				date_trunc('month', $4::date),
				'1 month'::interval
			)::date AS month_start
		),
		subscription_months AS (
			SELECT
				s.price,
				GREATEST(s.start_date, $3::date) AS effective_start,
				LEAST(COALESCE(s.end_date, $4::date), $4::date) AS effective_end
			FROM subscriptions s
			WHERE s.user_id = $1
			  AND s.service_name = $2
			  AND s.start_date <= $4::date
			  AND (s.end_date IS NULL OR s.end_date >= $3::date)
		)
		SELECT COALESCE(SUM(sm.price), 0)
		FROM subscription_months sm
		JOIN months_in_period m
			ON m.month_start BETWEEN sm.effective_start AND sm.effective_end
	`
