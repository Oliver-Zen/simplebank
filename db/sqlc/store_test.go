package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	// Initialize a new Store instance using the test database
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Println(">> before:", account1.Balance, account2.Balance)

	// run n concurrent transfer transactions (goroutines)
	n := 5
	amount := int64(10)

	errs := make(chan error)               // Channel to collect errors from goroutines
	results := make(chan TransferTxResult) // Channel to collect results from goroutines

	/* To debug potential DEADLOCK issues:
	- Print logs to trace which transaction is calling which query and in what order.
	- Use `context` to track and log transaction metadata, such as query flow and timing.
	- Ensure that locks (if any) are acquired in a consistent order to avoid circular waits.
	*/

	for i := 0; i < n; i++ {

		// for DEBUG:
		// txName := fmt.Sprintf("tx %d", i+1) // Create a unique transaction name for logging

		go func() {
			// Add transaction name to context for tracing and debugging
			// for DEBUG:
			// ctx := context.WithValue(context.Background(), txKey, txName)

			// The empty struct `txKey{}` creates a new instance,
			// but instances of the same type `txKey` are considered equal in Go.
			// This ensures that ctx.Value(txKey{}) can successfully retrieve the corresponding value.

			// Launch a goroutine for each concurrent transaction (different from `TestTransferTx`)
			// for DEBUG:
			// result, err := store.TransferTx(ctx, TransferTxParams{
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			errs <- err       // Send error (if any) to the errors channel
			results <- result // Send result to the results channel
		}()
	}

	// check results
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// check transfer
		existed := make(map[int]bool)

		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// check account entires
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		// check account's balance
		fmt.Println(">> tx: ", fromAccount.Balance, toAccount.Balance)
		diff1 := account1.Balance - fromAccount.Balance // money goes out
		diff2 := toAccount.Balance - account2.Balance   // moeny goes in
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0) // 1 * amont, 2 * amount, 3 * amount, ..., n * amount

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	// check the final updated balance
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	fmt.Println(">> after: ", updatedAccount1.Balance, updatedAccount2.Balance)
	require.Equal(t, account1.Balance-int64(n)*amount, updatedAccount1.Balance)
	require.Equal(t, account2.Balance+int64(n)*amount, updatedAccount2.Balance)

}

func TestTransferTxDeadlock(t *testing.T) {
	// Initialize a new Store instance using the test database
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Println(">> before:", account1.Balance, account2.Balance)

	// run n concurrent transfer transactions (goroutines)
	n := 10
	amount := int64(10)

	errs := make(chan error) // Channel to collect errors from goroutines
	// don't need to check `results`, as it already checked in `TestTransferTx`

	for i := 0; i < n; i++ {
		// half test cases: move money from account A -> account B
		// other half:      reverse, B -> A
		fromAccountID := account1.ID
		toAccountID := account2.ID

		if i%2 == 0 {
			toAccountID = account1.ID
			fromAccountID = account2.ID
		}

		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})

			errs <- err // Send error (if any) to the errors channel
		}()
	}

	// check results
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	// check the final updated balance
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	fmt.Println(">> after: ", updatedAccount1.Balance, updatedAccount2.Balance)
	require.Equal(t, account1.Balance, updatedAccount1.Balance)
	require.Equal(t, account2.Balance, updatedAccount2.Balance)

}
