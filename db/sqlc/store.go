package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
	Querier
}

// Store provides all functions to execute db queries and transactions
type SQLStore struct {
	*Queries
	db *pgxpool.Pool
}

// NewStore creates new store
func NewStore(db *pgxpool.Pool) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

// execTx execute a function within a db transaction
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}
	tx.Commit(ctx)
	return nil
}

// TransferTxParams contains the input parameters of the transfer transaction
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult is a result of transfer transaction
type TransferTxResult struct {
	Transfer    *Transfer `json:"transfer"`
	FromAccount *Account  `json:"from_account"`
	ToAccount   *Account  `json:"to_account"`
	FromEntry   *Entry    `json:"from_entry"`
	ToEntry     *Entry    `json:"to_entry"`
}

// TransferTx performs a money transfer from one account to the other,
// It creates a transfer record, add account entries, and update accounts balance within a single db tx
func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Transfer, err = q.CreateTransfer(ctx, &CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, &CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, &CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
		} else {
			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
		}

		return nil
	})

	return result, err
}

func addMoney(
	ctx context.Context,
	q *Queries,
	accID1 int64,
	amount1 int64,
	accID2 int64,
	amount2 int64,
) (acc1 *Account, acc2 *Account, err error) {
	acc1, err = q.AddAccountBalance(ctx, &AddAccountBalanceParams{
		ID:     accID1,
		Amount: amount1,
	})
	if err != nil {
		return
	}

	acc2, err = q.AddAccountBalance(ctx, &AddAccountBalanceParams{
		ID:     accID2,
		Amount: amount2,
	})
	return
}
