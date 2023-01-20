package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "github.com/the-eduardo/Go-Bank/db/mock"
	db "github.com/the-eduardo/Go-Bank/db/sqlc"
	"github.com/the-eduardo/Go-Bank/util"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func randomTransfer() db.Transfer {
	return db.Transfer{
		ID:            util.RandomInt(1, 1000),
		FromAccountID: util.RandomInt(1, 1000),
		ToAccountID:   util.RandomInt(1, 1000),
		Amount:        util.RandomInt(1, 1000),
	}
}

func requireBodyMatchTransfer(t *testing.T, body *bytes.Buffer, transfer db.Transfer) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	var gotTransfer db.Transfer
	err = json.Unmarshal(data, &gotTransfer)
	require.NoError(t, err)
	require.Equal(t, transfer, gotTransfer)
}

func requireBodyMatchTransferList(t *testing.T, body *bytes.Buffer, transfers []db.Transfer) {
	data, err := io.ReadAll(body)
	require.NoError(t, err, "error reading body")
	var gotTransfers []db.Transfer
	err = json.Unmarshal(data, &gotTransfers)
	require.NoError(t, err, "error unmarshalling body")
	require.Equal(t, transfers, gotTransfers, "transfers do not match")
}

func TestGetTransfer(t *testing.T) {
	transfer := randomTransfer()
	testCases := []struct {
		name          string
		transferID    int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:       "OK",
			transferID: transfer.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetTransfer(gomock.Any(), gomock.Eq(transfer.ID)).Times(1).Return(transfer, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code, "http.StatusOK") // check the response code
				requireBodyMatchTransfer(t, recorder.Body, transfer)            // check the response body
			},
		},
		{
			name:       "NotFound",
			transferID: transfer.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetTransfer(gomock.Any(), gomock.Eq(transfer.ID)).Times(1).Return(db.Transfer{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code, "Not Found") // check the response code
			},
		},
		{
			name:       "InternalError",
			transferID: transfer.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetTransfer(gomock.Any(), gomock.Eq(transfer.ID)).Times(1).Return(db.Transfer{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code, "server unreachable") // check the response code
			},
		},
		{
			name:       "InvalidID",
			transferID: -1,
			buildStubs: func(store *mockdb.MockStore) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code, "Invalid ID") // check the response code
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// create a mock store
			store := mockdb.NewMockStore(ctrl)
			// call the buildStubs function to setup the mock store
			tc.buildStubs(store)
			// create a server that uses the mock store
			server := NewServer(store)
			recorder := httptest.NewRecorder()
			// create a new request to our API
			myUrl := fmt.Sprintf("/transfer/%d", tc.transferID)
			request, err := http.NewRequest(http.MethodGet, myUrl, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListTransfers(t *testing.T) {
	n := 5
	transfers := make([]db.Transfer, n)
	for i := 0; i < n; i++ {
		transfers[i] = randomTransfer()
	}

	type QueryParams struct {
		FromAccountID int64 `json:"from_account_id"`
		ToAccountID   int64 `json:"to_account_id"`
		PageID        int32 `form:"page_id"`
		PageSize      int32 `form:"limit"`
	}
	testCases := []struct {
		name          string
		queryParams   QueryParams
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			queryParams: QueryParams{
				FromAccountID: transfers[0].FromAccountID,
				ToAccountID:   transfers[0].ToAccountID,
				PageSize:      5,
				PageID:        1,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListTransferParams{
					FromAccountID: transfers[0].FromAccountID,
					ToAccountID:   transfers[0].ToAccountID,
					Limit:         5,
					Offset:        1,
				}

				store.EXPECT().
					ListTransfer(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(transfers, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code, "http.StatusOK") // check the response code
				requireBodyMatchTransferList(t, recorder.Body, transfers)       // check the response body
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// create a mock store
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			myUrl := "/transfer"
			// create a new request to our API
			request, err := http.NewRequest(http.MethodGet, myUrl, nil)
			require.NoError(t, err)

			q := request.URL.Query()
			q.Add("from_account_id", strconv.FormatInt(tc.queryParams.FromAccountID, 10))
			q.Add("to_account_id", strconv.FormatInt(tc.queryParams.ToAccountID, 10))
			q.Add("page_id", strconv.FormatInt(int64(tc.queryParams.PageID), 10))
			q.Add("page_size", strconv.FormatInt(int64(tc.queryParams.PageSize), 10))
			request.URL.RawQuery = q.Encode()

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}
