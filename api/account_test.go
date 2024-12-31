package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/Oliver-Zen/simplebank/db/mock"
	db "github.com/Oliver-Zen/simplebank/db/sqlc"
	"github.com/Oliver-Zen/simplebank/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGetAccountAPI(t *testing.T) {
	account := randomAccount()

	// `testCases` is a slice of an anonymous struct
	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mockdb.MockStore) // GetAccount stub for each scenario will be built differently
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				// build stubs
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check response
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID, // WHY can use same ID? Because mock store is separated for each test case
			buildStubs: func(store *mockdb.MockStore) {
				// build stubs
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID, // WHY can use same ID? Because mock store is separated for each test case
			buildStubs: func(store *mockdb.MockStore) {
				// build stubs
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone) // WHAT ErrConnnDone? When a query is run on a connection that has been already returned to the connection pool
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				// build stubs
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0) // WHY 0? Because if ID is invalid, `GetAccount` should not be called
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			// defer ctrl.Finish() // new in go1.14+, no longer need to call ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			// build stubs
			// before `testCases` is created:
			// store.EXPECT().
			// GetAccount(gomock.Any(), gomock.Eq(account.ID)).
			// Times(1).
			// Return(account, nil)
			tc.buildStubs(store)

			// start test server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()
			// Creates an in-memory HTTP response recorder for testing HTTP handlers. 
			// It captures the handler's response (status code, headers, body) without starting a real HTTP server.

			// url := fmt.Sprintf("/accounts/%d", account.ID) // will causes [InvalidID] test case fails
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// send our API `request` through the server router
			// and record its response in the recorder
			server.router.ServeHTTP(recorder, request)
			// check response
			// before `testCases` is created
			// require.Equal(t, http.StatusOK, recorder.Code)
			// requireBodyMatchAccount(t, recorder.Body, account)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

// Because `recorder.Body` is just byte buffers.
// So create `requireBodyMatchAccount` to match the account.
func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}
