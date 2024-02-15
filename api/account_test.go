package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	mockdb "github.com/the-eduardo/Go-Bank/db/mock"
	db "github.com/the-eduardo/Go-Bank/db/sqlc"
	"github.com/the-eduardo/Go-Bank/util"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAccountAPI(t *testing.T) {
	account := randomAccount()

	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(mock.Anything, account.ID).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check the response
				assert.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body.Bytes(), account)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(mock.Anything, account.ID).
					Times(1).
					Return(db.Account{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(mock.Anything, account.ID).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		//{ // TODO: Fix this test someday
		//	name:      "InvalidID",
		//	accountID: 0,
		//	buildStubs: func(store *mockdb.MockStore) {
		//		store.EXPECT().
		//			GetAccount(mock.Anything, mock.Anything).
		//			Times(0)
		//	},
		//	checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		//		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		//	},
		//},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			mockStore := mockdb.NewMockStore(t)
			tc.buildStubs(mockStore)

			server := NewServer(mockStore)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func requireBodyMatchAccount(t *testing.T, body []byte, account db.Account) {
	var responseAccount db.Account
	err := json.Unmarshal(body, &responseAccount)
	assert.NoError(t, err)
	assert.Equal(t, account, responseAccount)
	assert.Equal(t, account.ID, responseAccount.ID)
	assert.Equal(t, account.Owner, responseAccount.Owner)
	assert.Equal(t, account.Balance, responseAccount.Balance)
	assert.Equal(t, account.Currency, responseAccount.Currency)
}

func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}
