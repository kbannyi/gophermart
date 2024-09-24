ALTER TABLE orders
    ALTER COLUMN accrual TYPE int USING accrual::integer;
ALTER TABLE withdrawals
    ALTER COLUMN amount TYPE int USING amount::integer;