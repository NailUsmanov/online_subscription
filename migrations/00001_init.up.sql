CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    service_name TEXT NOT NULL,
    price INT NOT NULL,
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE
);

-- Индекс для поиска по user_id 
CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);

-- Индекс для поиска по service_name
CREATE INDEX idx_subscriptions_service_name ON subscriptions(service_name);

-- Составной индекс для поиска по датам
CREATE INDEX idx_subscriptions_dates ON subscriptions(start_date, end_date);
