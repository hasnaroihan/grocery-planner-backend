package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Storage interface {
	Querier
	NewRecipeTx(ctx context.Context, arg NewRecipeParams) (RecipeResult, error)
	GetRecipeTx(ctx context.Context, id int64) (RecipeResult, error)
	UpdateRecipeTx(ctx context.Context, arg TxUpdateRecipeParams) (RecipeResult, error)
	GenerateGroceries(ctx context.Context, arg GenerateGroceriesParam) (GenerateGroceriesResult, error)
}

type SQLStorage struct {
	*Queries
	db *sql.DB
}

func NewStorage(db *sql.DB) *SQLStorage {
	return &SQLStorage{
		db: db,
		Queries: New(db),
	}
}

func (s *SQLStorage) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %v", err, rbErr)
		}

		return err
	}

	return tx.Commit()
}