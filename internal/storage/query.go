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
	WHERE user_id = $1;
`

// SumQuery - запрос для выдачи стоимости подписки пользователя за выбранный период.
var SumQuery string = `
		SELECT price, start_date, end_date
		FROM subscriptions
		WHERE user_id = $1
		  AND service_name = $2
		  AND start_date <= $4 
		  AND (end_date IS NULL OR end_date >= $3);
	`
