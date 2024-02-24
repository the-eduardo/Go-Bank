package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mockdb "github.com/the-eduardo/Go-Bank/db/mock"
	db "github.com/the-eduardo/Go-Bank/db/sqlc"
	"github.com/the-eduardo/Go-Bank/token"
	"github.com/the-eduardo/Go-Bank/util"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetEntryAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)
	entry := randomEntry(account.ID)

	testCases := []struct {
		name          string
		entryID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:    "OK",
			entryID: entry.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				account := db.Account{ID: entry.AccountID, Owner: user.Username}
				store.EXPECT().GetAccount(gomock.Any(), entry.AccountID).
					Times(1).Return(account, nil) // Mock GetAccount to return the account and nil error to pass validation
				store.EXPECT().GetEntry(gomock.Any(), entry.ID).Times(1).Return(entry, nil)
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntry(gomock.Any(), entry.ID).
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntry(gomock.Any(), entry.ID).
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// No call is expected because the request should fail validation.
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:    "UnauthorizedUser",
			entryID: entry.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, "unauthorized_user", user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEntry(gomock.Any(), gomock.Eq(entry.ID)).
					Times(0).
					Return(entry, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:    "NoAuthorization",
			entryID: entry.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEntry(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/entries/%d", tc.entryID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestCreateNewEntryAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)
	entry := randomEntry(account.ID)

	testCases := []struct {
		name          string
		accountID     int64
		Amount        int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "NewEntryCreated",
			accountID: entry.AccountID,
			Amount:    entry.Amount,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				account := db.Account{ID: entry.AccountID, Owner: user.Username}
				store.EXPECT().GetAccount(gomock.Any(), entry.AccountID).
					Times(1).Return(account, nil) // Mock GetAccount to return the account and nil error to pass validation
				store.EXPECT().AddAccountBalance(gomock.Any(), gomock.Eq(db.AddAccountBalanceParams{
					ID:     entry.AccountID,
					Amount: entry.Amount, // I'm not sure if I should do that
				})).Times(1).Return(db.Account{}, nil)
				store.EXPECT().NewEntry(gomock.Any(), gomock.Eq(db.NewEntryParams{
					AccountID: account.ID,
					Amount:    entry.Amount, // I'm not sure if I should do that
				})).Times(1).Return(entry, nil)
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				account := db.Account{ID: entry.AccountID, Owner: user.Username}
				store.EXPECT().GetAccount(gomock.Any(), entry.AccountID).
					Times(1).Return(account, nil) // Mock GetAccount to return the account and nil error to pass validation
				store.EXPECT().AddAccountBalance(gomock.Any(), gomock.Eq(db.AddAccountBalanceParams{
					ID:     entry.AccountID,
					Amount: entry.Amount,
				})).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "BadRequest",
			accountID: 0,
			Amount:    0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// No call is expected because the request should fail validation.
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "UnauthorizedUser",
			accountID: entry.AccountID,
			Amount:    entry.Amount,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, "unauthorized_user", user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account)).
					Times(0).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "NoAuthorization",
			accountID: entry.AccountID,
			Amount:    entry.Amount,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := "/entries"
			body := strings.NewReader(fmt.Sprintf(`{"account_id": %d, "amount": %d}`, tc.accountID, tc.Amount))
			request, err := http.NewRequest(http.MethodPost, url, body)
			assert.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListAccountEntriesAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)
	entry := randomEntry(account.ID)

	type Query struct {
		pageID    int
		pageSize  int
		AccountID int64
	}
	testCases := []struct {
		name          string
		query         Query
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				pageID:    1,
				pageSize:  5,
				AccountID: entry.AccountID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {

				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// Return a valid account with an owner
				account := db.Account{ID: entry.AccountID, Owner: user.Username}
				store.EXPECT().GetAccount(gomock.Any(), entry.AccountID).
					Times(1).Return(account, nil) // Mock GetAccount to return the account and nil error to pass validation

				store.EXPECT().ListEntries(gomock.Any(), db.ListEntriesParams{AccountID: entry.AccountID, Limit: 5, Offset: 0}).
					Times(1).
					Return([]db.Entry{entry}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check the response
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InternalError",
			query: Query{
				pageID:    1,
				pageSize:  5,
				AccountID: entry.AccountID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// Return a valid account with an owner
				account := db.Account{ID: entry.AccountID, Owner: user.Username}
				store.EXPECT().GetAccount(gomock.Any(), entry.AccountID).
					Times(1).Return(account, nil) // Mock GetAccount to pass validation
				store.EXPECT().ListEntries(gomock.Any(), db.ListEntriesParams{AccountID: entry.AccountID, Limit: 5, Offset: 0}).
					Times(1).
					Return([]db.Entry{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "NotFound",
			query: Query{
				pageID:    1,
				pageSize:  5,
				AccountID: entry.AccountID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), entry.AccountID).
					Times(1).
					Return(db.Account{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "BadRequest",
			query: Query{
				pageID:    0,
				pageSize:  0,
				AccountID: 0,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// No call is expected because the request should fail validation.
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "UnauthorizedUser",
			query: Query{
				pageID:    1,
				pageSize:  5,
				AccountID: entry.AccountID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, "unauthorized_user", user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account)).
					Times(0).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "NoAuthorization",
			query: Query{
				pageID:    1,
				pageSize:  5,
				AccountID: entry.AccountID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := "/entries/"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			q := request.URL.Query()
			q.Add("account_id", fmt.Sprintf("%d", tc.query.AccountID))
			q.Add("page_id", fmt.Sprintf("%d", tc.query.pageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			request.URL.RawQuery = q.Encode()
			//url := fmt.Sprintf("/entries/?account_id=%d&page_id=%d&page_size=%d", tc.AccountID, tc.PageID, tc.PageSize)

			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomEntry(accountID int64) db.Entry {
	return db.Entry{
		ID:        util.RandomInt(1, 1000),
		AccountID: accountID,
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
