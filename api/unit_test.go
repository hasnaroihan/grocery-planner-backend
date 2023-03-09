package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/hasnaroihan/grocery-planner/auth"
	dbmock "github.com/hasnaroihan/grocery-planner/db/mock"
	db "github.com/hasnaroihan/grocery-planner/db/sqlc"
	"github.com/hasnaroihan/grocery-planner/util"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestCreateUnitAPI(t *testing.T) {
	admin, _ := randomAdmin(t)
	unit := randomUnit()

	testCase := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker)
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"name": unit.Name,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					CreateUnit(gomock.Any(), gomock.Eq(unit.Name)).
					Times(1).
					Return(unit, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "400 Bad Name",
			body: gin.H{},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					CreateUnit(gomock.Any(), gomock.Eq(unit.Name)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "500 Connection Done",
			body: gin.H{
				"name":  unit.Name,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					CreateUnit(gomock.Any(), gomock.Eq(unit.Name)).
					Times(1).
					Return(db.Unit{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "409 Unique Violation",
			body: gin.H{
				"name": unit.Name,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					CreateUnit(gomock.Any(), gomock.Eq(unit.Name)).
					Times(1).
					Return(db.Unit{}, error(&pq.Error{
						Code: "23505",
					}))
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusConflict, recorder.Code)
			},
		},
	}

	for i := range testCase {
		tc := testCase[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storage := dbmock.NewMockStorage(ctrl)
			tc.buildStubs(storage)

			server := newTestServer(t, storage)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/unit/add"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))

			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestDeleteUnitAPI(t *testing.T) {
	admin, _ := randomAdmin(t)
	unit := randomUnit()

	testCases := []struct {
		name          string
		uri           int32
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker)
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			uri:  unit.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					DeleteUnit(gomock.Any(), gomock.Eq(unit.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "400 Invalid ID",
			uri:  -2,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					DeleteUnit(gomock.Any(), gomock.Eq(-2)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "404 Not Found",
			uri:  unit.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					DeleteUnit(gomock.Any(), gomock.Eq(unit.ID)).
					Times(1).
					Return(sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "500 Internal Server Error",
			uri:  unit.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					DeleteUnit(gomock.Any(), gomock.Eq(unit.ID)).
					Times(1).
					Return(sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storage := dbmock.NewMockStorage(ctrl)
			tc.buildStubs(storage)

			server := newTestServer(t, storage)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/unit/delete/%v", tc.uri)
			request, err := http.NewRequest(http.MethodDelete, url, nil)

			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetUnitAPI(t *testing.T) {
	admin, _ := randomUser(t)
	unit := randomUnit()

	testCases := []struct {
		name          string
		uri           int32
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker)
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			uri:  unit.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetUnit(gomock.Any(), gomock.Eq(unit.ID)).
					Times(1).
					Return(unit, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "400 Invalid ID",
			uri:  -2,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetUnit(gomock.Any(), gomock.Eq(-2)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "404 Not Found",
			uri: unit.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetUnit(gomock.Any(), gomock.Eq(unit.ID)).
					Times(1).
					Return(db.Unit{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "500 Internal Server Error",
			uri:  unit.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetUnit(gomock.Any(), gomock.Eq(unit.ID)).
					Times(1).
					Return(db.Unit{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storage := dbmock.NewMockStorage(ctrl)
			tc.buildStubs(storage)

			server := newTestServer(t, storage)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/unit/%v", tc.uri)
			request, err := http.NewRequest(http.MethodGet, url, nil)

			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListUnitAPI(t *testing.T) {
	units := []db.Unit{
		randomUnit(),
		randomUnit(),
		randomUnit(),
	}
	user, _ := randomUser(t)

	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker)
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					ListUnits(gomock.Any()).
					Times(1).
					Return(units, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "500 Internal Server Error",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					ListUnits(gomock.Any()).
					Times(1).
					Return([]db.Unit{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storage := dbmock.NewMockStorage(ctrl)
			tc.buildStubs(storage)

			server := newTestServer(t, storage)
			recorder := httptest.NewRecorder()

			url := "/unit/all"
			request, err := http.NewRequest(http.MethodGet, url, nil)

			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateUnitAPI(t *testing.T) {
	admin, _ := randomAdmin(t)
	unit := randomUnit()

	testCases := []struct {
		name          string
		uri           int32
		body          gin.H
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker)
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			uri: unit.ID,
			body: gin.H{
				"id":       unit.ID,
				"name": "new unit",
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.UpdateUnitParams{
					ID: unit.ID,
					Name: "new unit",
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					UpdateUnit(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(unit, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "400 Mismatched ID",
			uri: unit.ID,
			body: gin.H{
				"id":       unit.ID + 1,
				"name": "new unit",
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.UpdateUnitParams{
					ID: unit.ID,
					Name: "new unit",
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					UpdateUnit(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "400 Invalid UUID",
			uri: -unit.ID,
			body: gin.H{
				"id":       -unit.ID,
				"name": "new unit",
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.UpdateUnitParams{
					ID: -unit.ID,
					Name: "new unit",
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					UpdateUnit(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "404 Not Found",
			uri: unit.ID,
			body: gin.H{
				"id":       unit.ID,
				"name": "new unit",
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.UpdateUnitParams{
					ID: unit.ID,
					Name: "new unit",
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					UpdateUnit(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Unit{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "500 Internal Server Error",
			uri: unit.ID,
			body: gin.H{
				"id":       unit.ID,
				"name": "new unit",
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.UpdateUnitParams{
					ID: unit.ID,
					Name: "new unit",
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					UpdateUnit(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Unit{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storage := dbmock.NewMockStorage(ctrl)
			tc.buildStubs(storage)

			server := newTestServer(t, storage)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/unit/update/%v", tc.uri)
			request, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func randomUnit() db.Unit {
	return db.Unit{
		ID: int32(util.RandomInt(1,100)),
		Name: util.RandomUnit(),
	}
}