CREATE TABLE withdrawals
(
    id          varchar(100) NOT NULL,
    user_id     varchar(100) NOT NULL,
    order_id    varchar(100) NOT NULL,
    amount      int4         NOT NULL,
    created_utc timestamp    NOT NULL
);

CREATE UNIQUE INDEX withdrawals_order_id_uindex
    ON public.withdrawals (order_id);