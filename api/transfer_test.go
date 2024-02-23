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
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTransferAPI(t *testing.T) {
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
				store.EXPECT().GetTransferById(mock.Anything, transfer.ID).Times(1).Return(transfer, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check the response
				assert.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTransfer(t, recorder.Body.Bytes(), transfer)
			},
		},
		{
			name:       "NotFound",
			transferID: transfer.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetTransferById(mock.Anything, transfer.ID).
					Times(1).
					Return(db.Transfer{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:       "InternalError",
			transferID: transfer.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetTransferById(mock.Anything, transfer.ID).
					Times(1).
					Return(db.Transfer{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:       "InvalidID",
			transferID: 0,
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/transfers/%d", tc.transferID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

//func TestCreateNewTransferAPI(t *testing.T) { // TODO: TestCreateNewTransferAPI
//
//}

func randomTransfer() db.Transfer {
	return db.Transfer{
		ID:            util.RandomInt(1, 1000),
		ToAccountID:   util.RandomInt(501, 1000),
		FromAccountID: util.RandomInt(1, 500),
		Amount:        util.RandomMoney(),
	}
}
func requireBodyMatchTransfer(t *testing.T, body []byte, transfer db.Transfer) {
	var responseEntry db.Transfer
	err := json.Unmarshal(body, &responseEntry)
	assert.NoError(t, err)
	assert.Equal(t, transfer, responseEntry)
	assert.Equal(t, transfer.ID, responseEntry.ID)
	assert.Equal(t, transfer.FromAccountID, responseEntry.FromAccountID)
	assert.Equal(t, transfer.ToAccountID, responseEntry.ToAccountID)
	assert.Equal(t, transfer.Amount, responseEntry.Amount)

}
