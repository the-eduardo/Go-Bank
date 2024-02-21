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
	"strings"
	"testing"
)

func TestGetEntryAPI(t *testing.T) {
	entry := randomEntry()

	testCases := []struct {
		name          string
		entryID       int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:    "OK",
			entryID: entry.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntry(mock.Anything, entry.ID).Times(1).Return(entry, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check the response
				assert.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchEntry(t, recorder.Body.Bytes(), entry)
			},
		},
		{
			name:    "NotFound",
			entryID: entry.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntry(mock.Anything, entry.ID).
					Times(1).
					Return(db.Entry{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:    "InternalError",
			entryID: entry.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntry(mock.Anything, entry.ID).
					Times(1).
					Return(db.Entry{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:    "InvalidID",
			entryID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				// No call is expected because the request should fail validation.
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			mockStore := mockdb.NewMockStore(t)
			tc.buildStubs(mockStore)

			server := newTestServer(t, mockStore)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/entries/%d", tc.entryID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestCreateNewEntryAPI(t *testing.T) {
	entry := randomEntry()

	testCases := []struct {
		name          string
		accountID     int64
		Amount        int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "NewEntryCreated",
			accountID: entry.AccountID,
			Amount:    entry.Amount,

			buildStubs: func(store *mockdb.MockStore) {
				account := db.Account{ID: entry.AccountID} // Create a mock account
				store.EXPECT().GetAccount(mock.Anything, entry.AccountID).
					Times(1).Return(account, nil) // Mock GetAccount to return the account and nil error to pass validation
				store.EXPECT().AddAccountBalance(mock.Anything, db.AddAccountBalanceParams{ID: entry.AccountID, Amount: entry.Amount}).
					Times(1).
					Return(account, nil)
				store.EXPECT().NewEntry(mock.Anything, db.NewEntryParams{AccountID: entry.AccountID, Amount: entry.Amount}).
					Times(1).
					Return(entry, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check the response
				assert.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchEntry(t, recorder.Body.Bytes(), entry)
			},
		},
		{
			name:      "InternalError",
			accountID: entry.AccountID,
			Amount:    entry.Amount,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(mock.Anything, entry.AccountID).
					Times(1).Return(db.Account{}, nil) // Mock GetAccount to return the account and nil error to pass validation
				store.EXPECT().AddAccountBalance(mock.Anything, db.AddAccountBalanceParams{ID: entry.AccountID, Amount: entry.Amount}).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "BadRequest",
			accountID: 0,
			Amount:    0,
			buildStubs: func(store *mockdb.MockStore) {
				// No call is expected because the request should fail validation.
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			mockStore := mockdb.NewMockStore(t)
			tc.buildStubs(mockStore)

			server := newTestServer(t, mockStore)
			recorder := httptest.NewRecorder()

			url := "/entries"
			body := strings.NewReader(fmt.Sprintf(`{"account_id": %d, "amount": %d}`, tc.accountID, tc.Amount))
			request, err := http.NewRequest(http.MethodPost, url, body)
			assert.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListAccountEntriesAPI(t *testing.T) {
	entry := randomEntry()

	testCases := []struct {
		name          string
		AccountID     int64
		PageID        int64
		PageSize      int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			AccountID: entry.AccountID,
			PageID:    1,
			PageSize:  5,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(mock.Anything, entry.AccountID).
					Times(1).Return(db.Account{}, nil) // Mock GetAccount to pass validation
				store.EXPECT().ListEntries(mock.Anything, db.ListEntriesParams{AccountID: entry.AccountID, Limit: 5, Offset: 0}).
					Times(1).
					Return([]db.Entry{entry}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check the response
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			AccountID: entry.AccountID,
			PageID:    1,
			PageSize:  5,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(mock.Anything, entry.AccountID).
					Times(1).Return(db.Account{}, nil) // Mock GetAccount to pass validation
				store.EXPECT().ListEntries(mock.Anything, db.ListEntriesParams{AccountID: entry.AccountID, Limit: 5, Offset: 0}).
					Times(1).
					Return([]db.Entry{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "NotFound",
			AccountID: entry.AccountID,
			PageID:    1,
			PageSize:  5,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(mock.Anything, entry.AccountID).
					Times(1).
					Return(db.Account{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "BadRequest",
			AccountID: 0,
			PageID:    0,
			PageSize:  0,
			buildStubs: func(store *mockdb.MockStore) {
				// No call is expected because the request should fail validation.
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			mockStore := mockdb.NewMockStore(t)
			tc.buildStubs(mockStore)

			server := newTestServer(t, mockStore)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/entries/?account_id=%d&page_id=%d&page_size=%d", tc.AccountID, tc.PageID, tc.PageSize)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomEntry() db.Entry {
	return db.Entry{
		ID:        util.RandomInt(1, 1000),
		AccountID: util.RandomInt(1, 1000),
		Amount:    util.RandomMoney(),
	}
}
func requireBodyMatchEntry(t *testing.T, body []byte, entry db.Entry) {
	var responseEntry db.Entry
	err := json.Unmarshal(body, &responseEntry)
	assert.NoError(t, err)
	assert.Equal(t, entry, responseEntry)
	assert.Equal(t, entry.ID, responseEntry.ID)
	assert.Equal(t, entry.AccountID, responseEntry.AccountID)
	assert.Equal(t, entry.Amount, responseEntry.Amount)
}
