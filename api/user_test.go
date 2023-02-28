package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	dbmock "github.com/hasnaroihan/grocery-planner/db/mock"
	db "github.com/hasnaroihan/grocery-planner/db/sqlc"
	"github.com/hasnaroihan/grocery-planner/util"
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
		ID: id,
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
		ID: id,
		Username: util.RandomUsername(),
		Password: hashedPass,
		Email:    util.RandomEmail(),
		Role:     "admin",
	}, password
}
