package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/kbannyi/gophermart/internal/domain"
)

type OrderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) OrderRepository {
	return OrderRepository{db}
}

func (r OrderRepository) SaveNewOrder(ctx context.Context, order domain.Order) error {
	_, err := r.db.NamedExecContext(ctx,
		`INSERT INTO orders (id, status, user_id, created_utc, updated_utc) 
		VALUES (:id, :status, :user_id, :created_utc, :updated_utc)`,
		order)
	if err != nil {
		return convertErr(err)
	}

	return nil
}

func (r OrderRepository) Get(ctx context.Context, id string) (*domain.Order, error) {
	var o domain.Order
	err := r.db.GetContext(ctx, &o, "SELECT * FROM orders WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &o, nil
}

func (r OrderRepository) GetOrders(ctx context.Context, userid string) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.SelectContext(ctx, &orders, "SELECT * FROM orders WHERE user_id = $1 ORDER BY created_utc DESC", userid)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r OrderRepository) SelectForFetching(ctx context.Context, orders *[]domain.Order, pageSize int, page int) error {
	offset := page * pageSize
	err := r.db.SelectContext(ctx,
		orders,
		`SELECT *
			FROM orders
			WHERE status = $3 OR status = $4
			ORDER BY created_utc
			LIMIT $1 OFFSET $2;`,
		pageSize, offset, domain.StatusNew.String(), domain.StatusProcessing.String())
	if err != nil {
		return err
	}

	return nil
}

func (r OrderRepository) BatchSave(ctx context.Context, orders []domain.Order) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.NamedExecContext(ctx, `
	UPDATE orders
	SET status=:status,
		accrual=:accrual,
		updated_utc=:updated_utc
	WHERE id = :id;
	`, orders)
	if err != nil {
		return err
	}

	return tx.Commit()
}
