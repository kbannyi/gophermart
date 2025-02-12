package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/kbannyi/gophermart/internal/domain"
	"github.com/kbannyi/gophermart/internal/models"
	"github.com/shopspring/decimal"
)

type WithdrawalRepository struct {
	db *sqlx.DB
}

func NewWithdrawalRepository(db *sqlx.DB) WithdrawalRepository {
	return WithdrawalRepository{
		db: db,
	}
}

var ErrNotEnoughPoints = errors.New("not enough points")

func (r WithdrawalRepository) Withdraw(ctx context.Context, w domain.Withdrawal) error {
	var accruals []decimal.Decimal
	err := r.db.SelectContext(ctx, &accruals, `
	SELECT accrual
	FROM orders
	WHERE user_id = $1
	AND accrual IS NOT NULL;
	`, w.UserID)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var withdrawals []decimal.Decimal
	err = tx.SelectContext(ctx, &withdrawals, `
	SELECT amount
	FROM withdrawals
	WHERE user_id = $1;
	`, w.UserID)
	if err != nil {
		return err
	}
	earned := decimal.Sum(decimal.Zero, accruals...)
	withdrawn := decimal.Sum(w.Amount, withdrawals...)
	if earned.LessThan(withdrawn) {
		return ErrNotEnoughPoints
	}
	_, err = tx.NamedExecContext(ctx, `
	INSERT INTO withdrawals(id, user_id, order_id, amount, created_utc)
	VALUES (:id, :user_id, :order_id, :amount, :created_utc);
	`, w)
	if err != nil {
		return convertErr(err)
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (r WithdrawalRepository) GetBalance(ctx context.Context, userid string) (*models.Balance, error) {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var earned []decimal.Decimal
	err = tx.SelectContext(ctx, &earned, `
	SELECT accrual
	FROM orders
	WHERE user_id = $1
	AND accrual IS NOT NULL;
	`, userid)
	if err != nil {
		return nil, err
	}
	var withdrawn []decimal.Decimal
	err = tx.SelectContext(ctx, &withdrawn, `
	SELECT amount
	FROM withdrawals
	WHERE user_id = $1;
	`, userid)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	totalEarned := decimal.Sum(decimal.Zero, earned...)
	totalWithdrawn := decimal.Sum(decimal.Zero, withdrawn...)
	balance := models.Balance{
		UserID:    userid,
		Current:   totalEarned.Sub(totalWithdrawn),
		Withdrawn: totalWithdrawn,
	}

	return &balance, nil
}

func (r WithdrawalRepository) GetWithdrawals(ctx context.Context, userID string) ([]domain.Withdrawal, error) {
	var withdrawals []domain.Withdrawal
	err := r.db.SelectContext(ctx, &withdrawals, "SELECT * FROM withdrawals WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}
