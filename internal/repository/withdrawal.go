package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/kbannyi/gophermart/internal/domain"
	"github.com/kbannyi/gophermart/internal/models"
)

type WithdrawalRepository struct {
	db *sqlx.DB
}

func NewWithdrawalRepository(db *sqlx.DB) WithdrawalRepository {
	return WithdrawalRepository{
		db: db,
	}
}

var ErrNotEnoughPoints = errors.New("not enough points on the balance")

func (r WithdrawalRepository) Withdraw(ctx context.Context, w domain.Withdrawal) error {
	earned := new(int)
	err := r.db.GetContext(ctx, &earned, `
	SELECT COALESCE(SUM(accrual), 0)
	FROM orders
	WHERE user_id = $1
	AND accrual IS NOT NULL;
	`, w.UserId)
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

	withdrawn := new(int)
	err = tx.GetContext(ctx, withdrawn, `
	SELECT COALESCE(SUM(amount), 0)
	FROM withdrawals
	WHERE user_id = $1;
	`, w.UserId)
	if err != nil {
		return err
	}
	if *earned < (*withdrawn + w.Amount) {
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

	balance := models.Balance{UserID: userid}
	earned := new(int)
	err = tx.GetContext(ctx, &earned, `
	SELECT COALESCE(SUM(accrual), 0)
	FROM orders
	WHERE user_id = $1
	AND accrual IS NOT NULL;
	`, userid)
	if err != nil {
		return nil, err
	}
	err = tx.GetContext(ctx, &balance.Withdrawn, `
	SELECT COALESCE(SUM(amount), 0)
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

	balance.Current = *earned - balance.Withdrawn

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
