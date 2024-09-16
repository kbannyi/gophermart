CREATE TABLE orders
(
    id          varchar(100) PRIMARY KEY,
    status      varchar(100)                       NOT NULL,
    user_id     varchar(100) REFERENCES users (id) NOT NULL,
    accrual     integer                            NULL,
    created_utc timestamp                          NOT NULL,
    updated_utc timestamp                          NULL
);

CREATE INDEX orders_status_index
    ON public.orders (status);