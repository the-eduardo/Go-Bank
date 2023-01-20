package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "github.com/the-eduardo/Go-Bank/db/mock"
	db "github.com/the-eduardo/Go-Bank/db/sqlc"
	"github.com/the-eduardo/Go-Bank/util"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func randomEntry() db.Entry {
	return db.Entry{
		ID:        util.RandomInt(1, 1000),
		AccountID: util.RandomInt(1, 1000),
		Amount:    util.RandomBalance(),
	}
}

func requireBodyMatchEntry(t *testing.T, body io.Reader, entry db.Entry) {
	data, err := io.ReadAll(body)
	require.NoError(t, err, "read body")
	var gotEntry db.Entry
	err = json.Unmarshal(data, &gotEntry)
	require.NoError(t, err, "unmarshal entry")
	require.Equal(t, entry, gotEntry, "entry")
}
func requireBodyMatchEntries(t *testing.T, body io.Reader, entry []db.Entry) {
	data, err := io.ReadAll(body)
	require.NoError(t, err, "read body")
	var got db.Entry
	err = json.Unmarshal(data, &got)
	require.NoError(t, err, "unmarshal entry")
	require.Equal(t, entry, got, "entry")
}

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
				store.EXPECT().GetEntries(gomock.Any(), gomock.Eq(entry.ID)).Times(1).Return(entry, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code, "http.StatusOK") // check the response code
				requireBodyMatchEntry(t, recorder.Body, entry)                  // check the response body
			},
		},
		{
			name:    "InternalError",
			entryID: entry.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntries(gomock.Any(), gomock.Eq(entry.ID)).Times(1).Return(db.Entry{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code, "server unreachable") // check the response code
			},
		},
		{
			name:    "InvalidID",
			entryID: -1,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntries(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code, "invalid id") // check the response code
			},
		},
	}
	for i := range testCases {
		testCase := testCases[i]
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			testCase.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()
			myUrl := fmt.Sprintf("/entries/%d", testCase.entryID)
			request, err := http.NewRequest(http.MethodGet, myUrl, nil)
			require.NoError(t, err, "request")

			server.router.ServeHTTP(recorder, request)
			testCase.checkResponse(t, recorder)
		})
	}
}

func TestCreateNewEntry(t *testing.T) {
	entry := randomEntry()
	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"account_id": entry.AccountID,
				"amount":     entry.Amount,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateEntriesParams{
					AccountID: entry.AccountID,
					Amount:    entry.Amount,
				}
				store.EXPECT().CreateEntries(gomock.Any(), gomock.Eq(arg)).Times(1).Return(entry, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code, "http.StatusOK") // check the response code
				requireBodyMatchEntry(t, recorder.Body, entry)                  // check the response body
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"account_id": entry.AccountID,
				"amount":     entry.Amount,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateEntriesParams{
					AccountID: entry.AccountID,
					Amount:    entry.Amount,
				}
				store.EXPECT().CreateEntries(gomock.Any(), gomock.Eq(arg)).Times(1).Return(db.Entry{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code, "server unreachable") // check the response code
			},
		},
		{
			name: "InvalidID",
			body: gin.H{
				"account_id": -1,
				"amount":     entry.Amount,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateEntriesParams{
					AccountID: -1,
					Amount:    entry.Amount,
				}
				store.EXPECT().CreateEntries(gomock.Any(), gomock.Eq(arg)).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code, "invalid id") // check the response code
			},
		},
		{
			name: "InvalidAmount",
			body: gin.H{
				"account_id": entry.AccountID,
				"amount":     -1,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateEntriesParams{
					AccountID: entry.AccountID,
					Amount:    -1,
				}
				store.EXPECT().CreateEntries(gomock.Any(), gomock.Eq(arg)).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code, "invalid id") // check the response code
			},
		},
	}
	for i := range testCases {
		testCase := testCases[i]
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			testCase.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			requestBody, err := json.Marshal(testCase.body)
			require.NoError(t, err, "request body")

			myUrl := fmt.Sprintf("/entries")
			request, err := http.NewRequest(http.MethodPost, myUrl, bytes.NewBuffer(requestBody))
			require.NoError(t, err, "request")

			server.router.ServeHTTP(recorder, request)
			testCase.checkResponse(t, recorder)
		})
	}
}

func TestListEntries(t *testing.T) {
	n := 5
	entries := make([]db.Entry, n)
	for i := 0; i < n; i++ {
		entries[i] = randomEntry()
	}

	type Query struct {
		AccountID int64 `form:"account_id"`
		Limit     int   `form:"limit"`
		Offset    int   `form:"offset"`
	}

	testCases := []struct {
		name          string
		query         Query
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				AccountID: entries[0].AccountID,
				Limit:     5,
				Offset:    1,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListEntriesParams{
					AccountID: entries[0].AccountID,
					Limit:     5,
					Offset:    0,
				}
				store.EXPECT().
					ListEntries(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(entries, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code, "http.StatusOK") // check the response code
				requireBodyMatchEntries(t, recorder.Body, entries)              // check the response body
			},
		},
		{
			name: "InternalError",
			query: Query{
				AccountID: entries[0].AccountID,
				Limit:     5,
				Offset:    1,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListEntriesParams{
					AccountID: entries[0].AccountID,
					Limit:     5,
					Offset:    1,
				}
				store.EXPECT().ListEntries(gomock.Any(), gomock.Eq(arg)).Times(1).Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code, "server unreachable") // check the response code
			},
		},
		{
			name: "InvalidID",
			query: Query{
				AccountID: -1,
				Limit:     5,
				Offset:    1,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListEntriesParams{
					AccountID: -1,
					Limit:     5,
					Offset:    1,
				}
				store.EXPECT().ListEntries(gomock.Any(), gomock.Eq(arg)).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code, "invalid id") // check the response code
			},
		},
		{
			name: "InvalidLimit",
			query: Query{
				AccountID: entries[0].AccountID,
				Limit:     -1,
				Offset:    0,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListEntriesParams{
					AccountID: entries[0].AccountID,
					Limit:     -1,
					Offset:    1,
				}
				store.EXPECT().ListEntries(gomock.Any(), gomock.Eq(arg)).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code, "invalid id") // check the response code
			},
		},
		{
			name: "InvalidOffset",
			query: Query{
				AccountID: entries[0].AccountID,
				Limit:     n,
				Offset:    -1,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListEntriesParams{
					AccountID: entries[0].AccountID,
					Limit:     5,
					Offset:    -1,
				}
				store.EXPECT().GetEntries(gomock.Any(), gomock.Eq(arg)).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code, "invalid id") // check the response code
			},
		},
	}
	for i := range testCases {
		testCase := testCases[i]
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			testCase.buildStubs(store)
			server := NewServer(store)
			recorder := httptest.NewRecorder()
			url := fmt.Sprintf("/entries?account_id=%d&page_id=%d&page_size=%d", testCase.query.AccountID, testCase.query.Limit, testCase.query.Offset)
			request, _ := http.NewRequest(http.MethodGet, url, nil)
			server.router.ServeHTTP(recorder, request)
			testCase.checkResponse(t, recorder)
		})
	}
}
