package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/hasnaroihan/grocery-planner/auth"
	dbmock "github.com/hasnaroihan/grocery-planner/db/mock"
	db "github.com/hasnaroihan/grocery-planner/db/sqlc"
	"github.com/hasnaroihan/grocery-planner/util"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestRegisterAPI(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username": user.Username,
				"email":    user.Email,
				"password": password,
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.CreateUserParams{
					Username: user.Username,
					Email:    user.Email,
					Role:     user.Role,
				}
				storage.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "400 Invalid Username",
			body: gin.H{
				"username": "randomuserM<script></script>",
				"email":    user.Email,
				"password": password,
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.CreateUserParams{
					Username: "randomuserM<script></script>",
					Email:    user.Email,
					Role:     user.Role,
				}
				storage.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "400 Invalid Email",
			body: gin.H{
				"username": user.Username,
				"email":    user.Username,
				"password": password,
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.CreateUserParams{
					Username: user.Username,
					Email:    user.Username,
					Role:     user.Role,
				}
				storage.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "400 Password Too Short",
			body: gin.H{
				"username": user.Username,
				"email":    user.Email,
				"password": "pass",
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "500 Password Too Long",
			body: gin.H{
				"username": user.Username,
				"email":    user.Email,
				"password": util.RandomString(100),
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "500 Connection Done",
			body: gin.H{
				"username": user.Username,
				"email":    user.Email,
				"password": password,
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.CreateUserParams{
					Username: user.Username,
					Email:    user.Email,
					Role:     user.Role,
				}
				storage.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "500 Unique Violation",
			body: gin.H{
				"username": user.Username,
				"email":    user.Email,
				"password": password,
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.CreateUserParams{
					Username: user.Username,
					Email:    user.Email,
					Role:     user.Role,
				}
				storage.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
					Times(1).
					Return(db.User{}, error(&pq.Error{
						Code: "23505",
					}))
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusConflict, recorder.Code)
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

			url := "/register"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestLoginAPI(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username": user.Username,
				"password": password,
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetLogin(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "400 Invalid Username",
			body: gin.H{
				"username": "randomuserM<script></script>",
				"password": password,
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetLogin(gomock.Any(), gomock.Eq("randomuserM<script></script>")).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "400 Password Too Short",
			body: gin.H{
				"username": user.Username,
				"password": "pass",
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetLogin(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "404 User Not Found",
			body: gin.H{
				"username": user.Username,
				"password": user.Password,
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetLogin(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "500 Internal Server Error",
			body: gin.H{
				"username": user.Username,
				"password": user.Password,
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetLogin(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "401 Wrong Password",
			body: gin.H{
				"username": user.Username,
				"password": "wrongpassword",
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetLogin(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
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

			url := "/login"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestDeleteUserAPI(t *testing.T) {
	user, _ := randomUser(t)
	admin, _ := randomAdmin(t)

	testCases := []struct {
		name          string
		uri           string
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker)
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			uri:  user.ID.String(),
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
					DeleteUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "400 Invalid UUID",
			uri:  "uu23-2342ec",
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
					DeleteUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "404 Not Found",
			uri:  user.ID.String(),
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
					DeleteUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).Return(sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "500 Internal Server Error",
			uri:  user.ID.String(),
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
					DeleteUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).Return(sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "403 Not Admin",
			uri:  user.ID.String(),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "common",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					DeleteUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
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

			url := fmt.Sprintf("/user/delete/%s", tc.uri)
			request, err := http.NewRequest(http.MethodDelete, url, nil)

			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListUsersAPI(t *testing.T) {
	var users []db.User
	admin, _ := randomAdmin(t)

	testCases := []struct {
		name          string
		query         string
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker)
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: "pageSize=2&pageNum=1",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.ListUsersParams{
					Limit:  2,
					Offset: 0,
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					ListUsers(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(users, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:  "400 Bad Request",
			query: "pageSize=2",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.ListUsersParams{
					Limit:  2,
					Offset: 0,
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					ListUsers(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:  "500 Internal Server Error",
			query: "pageSize=2&pageNum=1",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.ListUsersParams{
					Limit:  2,
					Offset: 0,
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					ListUsers(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil, sql.ErrConnDone)
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

			url := fmt.Sprintf("/user/all?%s", tc.query)
			request, err := http.NewRequest(http.MethodGet, url, nil)

			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateUserAPI(t *testing.T) {
	user, _ := randomUser(t)
	admin, _ := randomAdmin(t)

	testCases := []struct {
		name          string
		uri           string
		body          gin.H
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker)
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK User",
			uri: user.ID.String(),
			body: gin.H{
				"id":       user.ID,
				"username": "new_username",
				"email":    "new@grocery-planner.com",
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.UpdateUserParams{
					ID:       user.ID,
					Username: "new_username",
					Email:    "new@grocery-planner.com",
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "common",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.User{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "OK Admin",
			uri: user.ID.String(),
			body: gin.H{
				"id":       user.ID,
				"username": "new_username",
				"email":    "new@grocery-planner.com",
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.UpdateUserParams{
					ID:       user.ID,
					Username: "new_username",
					Email:    "new@grocery-planner.com",
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.User{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "400 Mismatched UUID",
			uri: user.ID.String(),
			body: gin.H{
				"id":       admin.ID,
				"username": "new_username",
				"email":    "new@grocery-planner.com",
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.UpdateUserParams{
					ID:       admin.ID,
					Username: "new_username",
					Email:    "new@grocery-planner.com",
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(0)
				storage.EXPECT().
					UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "400 Invalid URI",
			uri: "fff-0000",
			body: gin.H{
				"id":       user.ID,
				"username": "new_username",
				"email":    "new@grocery-planner.com",
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.UpdateUserParams{
					ID:       user.ID,
					Username: "new_username",
					Email:    "new@grocery-planner.com",
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(0)
				storage.EXPECT().
					UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "400 Invalid UUID",
			uri: user.ID.String(),
			body: gin.H{
				"id":       "ffff-0000",
				"username": "new_username",
				"email":    "new@grocery-planner.com",
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.UpdateUserParams{
					ID:       user.ID,
					Username: "new_username",
					Email:    "new@grocery-planner.com",
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(0)
				storage.EXPECT().
					UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "403 Forbidden",
			uri: user.ID.String(),
			body: gin.H{
				"id":       user.ID,
				"username": "new_username",
				"email":    "new@grocery-planner.com",
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				id,_ := uuid.NewRandom()
				addAuthorization(t, req, tokenMaker, authBearerType, id, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.UpdateUserParams{
					ID:       user.ID,
					Username: "new_username",
					Email:    "new@grocery-planner.com",
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.GetPermissionRow{
						Role: "common",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "404 Not Found",
			uri: user.ID.String(),
			body: gin.H{
				"id":       user.ID,
				"username": "new_username",
				"email":    "new@grocery-planner.com",
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.UpdateUserParams{
					ID:       user.ID,
					Username: "new_username",
					Email:    "new@grocery-planner.com",
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "500 Internal Server Error",
			uri: user.ID.String(),
			body: gin.H{
				"id":       user.ID,
				"username": "new_username",
				"email":    "new@grocery-planner.com",
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.UpdateUserParams{
					ID:       user.ID,
					Username: "new_username",
					Email:    "new@grocery-planner.com",
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
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

			url := fmt.Sprintf("/user/update/%s", tc.uri)
			request, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetUserAPI(t *testing.T) {
	user,_ := randomUser(t)
	admin,_ := randomAdmin(t)

	testCases := []struct {
		name          string
		uri           string
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker)
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK User",
			uri: user.ID.String(),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
				GetPermission(gomock.Any(), gomock.Eq(user.ID)).
				Times(1).
				Return(db.GetPermissionRow{
					Role: "common",
				}, nil)
				storage.EXPECT().
				GetUser(gomock.Any(), gomock.Eq(user.ID)).
				Times(1).
				Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "OK Admin",
			uri: user.ID.String(),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
				GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
				Times(1).
				Return(db.GetPermissionRow{
					Role: "admin",
				}, nil)
				storage.EXPECT().
				GetUser(gomock.Any(), gomock.Eq(user.ID)).
				Times(1).
				Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "400 Invalid UUID",
			uri: "ffff-0000",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
				GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
				Times(0)
				storage.EXPECT().
				GetUser(gomock.Any(), gomock.Any()).
				Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "403 Forbidden",
			uri: user.ID.String(),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				id,_ := uuid.NewRandom()
				addAuthorization(t, req, tokenMaker, authBearerType, id, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
				GetPermission(gomock.Any(), gomock.Any()).
				Times(1).
				Return(db.GetPermissionRow{
					Role: "common",
				}, nil)
				storage.EXPECT().
				GetUser(gomock.Any(), gomock.Eq(user.ID)).
				Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "404 Not Found",
			uri: user.ID.String(),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
				GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
				Times(1).
				Return(db.GetPermissionRow{
					Role: "admin",
				}, nil)
				storage.EXPECT().
				GetUser(gomock.Any(), gomock.Eq(user.ID)).
				Times(1).
				Return(db.User{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "500 Internal Server Error",
			uri: user.ID.String(),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
				GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
				Times(1).
				Return(db.GetPermissionRow{
					Role: "admin",
				}, nil)
				storage.EXPECT().
				GetUser(gomock.Any(), gomock.Eq(user.ID)).
				Times(1).
				Return(db.User{}, sql.ErrConnDone)
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

			url := fmt.Sprintf("/user/%s", tc.uri)
			request, err := http.NewRequest(http.MethodGet, url, nil)

			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}
type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	err := util.ComparePassword(e.password, arg.Password)
	if err != nil {
		return false
	}

	e.arg.Password = arg.Password
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password}
}

func randomUser(t *testing.T) (db.User, string) {
	password := util.RandomString(8)
	hashedPass, err := util.HashPassword(password)
	require.NoError(t, err)

	id, err := uuid.NewRandom()
	require.NoError(t, err)

	return db.User{
		ID:       id,
		Username: util.RandomUsername(),
		Password: hashedPass,
		Email:    util.RandomEmail(),
		Role:     "common",
	}, password
}

func randomAdmin(t *testing.T) (db.User, string) {
	password := util.RandomString(8)
	hashedPass, err := util.HashPassword(password)
	require.NoError(t, err)

	id, err := uuid.NewRandom()
	require.NoError(t, err)

	return db.User{
		ID:       id,
		Username: util.RandomUsername(),
		Password: hashedPass,
		Email:    util.RandomEmail(),
		Role:     "admin",
	}, password
}
