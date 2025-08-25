package db

import (
	"errors"

	"github.com/mishkahtherapy/brain/core/ports"
)

var ErrFailedToUpdate = errors.New("failed to update")

type SQLTransactionRepo struct {
	db ports.SQLDatabase
}

func NewSQLTransactionRepo(db ports.SQLDatabase) ports.TransactionPort {
	return &SQLTransactionRepo{db: db}
}

func (r *SQLTransactionRepo) Begin() (ports.SQLTx, error) {
	return r.db.Begin()
}

func (r *SQLTransactionRepo) Commit(tx ports.SQLTx) error {
	err := tx.Commit()
	if err != nil {
		return ErrFailedToUpdate
	}
	return nil
}

func (r *SQLTransactionRepo) Rollback(tx ports.SQLTx) error {
	err := tx.Rollback()
	if err != nil {
		return ErrFailedToUpdate
	}
	return nil
}
