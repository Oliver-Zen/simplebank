package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provides all functions to execute db queries and transaction
type Store struct {
	*Queries // Struct Embedding
	db       *sql.DB
}

// NewStore creates a new store
func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

// ExecTx executes a function within a database transaction
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {

	tx, err := store.db.BeginTx(ctx, nil) // Start a database transaction with the provided context
	if err != nil {
		return err
	}

	q := New(tx) // Create a new Queries instance bound to the transaction

	err = fn(q) // Execute the provided function `fn` using the transaction-bound Queries
	if err != nil {
		// rb - rollback
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}
	return tx.Commit() // Commit the transaction if `fn` succeeds
}

// TransferTxParams contains the input parameters of the transfer transaction
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult is the result of the transfer transaction
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// for DEBUG:
// var txKey = struct{}{} // txKey is a unique context key for transaction metadata.

// TransferTx performs a money transfer from one account to the other.
// It creates the transfer, add account entries, and update accounts' balance within a database transaction
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult // Initialize the result structure to store transaction details

	// Execute the transaction logic within a database transaction
	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// for DEBUG:
		// txName := ctx.Value(txKey) // Extract transaction name from the context for logging

		// for DEBUG:
		// fmt.Println(txName, "create transfer")
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		// for DEBUG:
		// fmt.Println(txName, "create entry 1")
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		// for DEBUG:
		// fmt.Println(txName, "create entry 2")
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		// To avoid deadlocks, the order of operations is critical.
		// Always lock rows in a consistent order based on account IDs.
		// If FromAccountID < ToAccountID, process FromAccount first, then ToAccount.
		// Otherwise, process ToAccount first, then FromAccount.
		// This ensures that transactions accessing the same rows acquire locks in the same order, avoiding circular waits.
		if arg.FromAccountID < arg.ToAccountID {
			/* before refactor (old version)
			// before AddAccountBalance (NO KEY UPDATE), old version
			// for DEBUG:
			// fmt.Println(txName, "get account 1")
			// account1, err := q.GetAccountForUpdate(ctx, arg.FromAccountID)
			// if err != nil {
			// 	return err
			// }

			// for DEBUG:
			// fmt.Println(txName, "update account 1")
			// result.FromAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
			// 	ID:      arg.FromAccountID,
			// 	Balance: account1.Balance - arg.Amount,
			// })
			// if err != nil {
			// 	return err
			// }
			result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     arg.FromAccountID,
				Amount: -arg.Amount,
			})
			if err != nil {
				return err
			}

			// before AddAccountBalance (NO KEY UPDATE). old version
			// for DEBUG:
			// fmt.Println(txName, "get account 2")
			// account2, err := q.GetAccountForUpdate(ctx, arg.ToAccountID)
			// if err != nil {
			// 	return err
			// }

			// for DEBUG:
			// fmt.Println(txName, "update account 2")
			// result.ToAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
			// 	ID:      arg.ToAccountID,
			// 	Balance: account2.Balance + arg.Amount,
			// })
			// if err != nil {
			// 	return err
			// }
			result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     arg.ToAccountID,
				Amount: arg.Amount,
			})
			if err != nil {
				return err
			} */
			result.FromAccount, result.ToAccount, err =
				addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
		} else {
			/* before refactor (old version)
			result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     arg.ToAccountID,
				Amount: arg.Amount,
			})
			if err != nil {
				return err
			}

			result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     arg.FromAccountID,
				Amount: -arg.Amount,
			})
			if err != nil {
				return err
			} */
			result.ToAccount, result.FromAccount, err =
				addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
		}

		return nil
	})

	// Return the transaction result and any error encountered
	return result, err
}

func addMoney(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64,
) (account1 Account, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return
		// same as:
		// return account1, account2, err
		// syntax feature of Go for conciseness
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})
	return
}
