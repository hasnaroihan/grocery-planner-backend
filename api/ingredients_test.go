package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/hasnaroihan/grocery-planner/auth"
	dbmock "github.com/hasnaroihan/grocery-planner/db/mock"
	db "github.com/hasnaroihan/grocery-planner/db/sqlc"
	"github.com/hasnaroihan/grocery-planner/util"
	"github.com/stretchr/testify/require"
)

func TestCreateIngredientAPI(t *testing.T) {
	admin, _ := randomAdmin(t)
	ingredient := randomIngredient(t)

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
				"name":        strings.ToLower(ingredient.Name),
				"defaultUnit": ingredient.DefaultUnit,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.CreateIngredientParams{
					Name:        strings.ToLower(ingredient.Name),
					DefaultUnit: ingredient.DefaultUnit,
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					CreateIngredient(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(ingredient, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "400 Bad Name",
			body: gin.H{
				"name":        ingredient.Name,
				"defaultUnit": ingredient.DefaultUnit,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.CreateIngredientParams{
					Name:        strings.ToLower(ingredient.Name),
					DefaultUnit: ingredient.DefaultUnit,
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					CreateIngredient(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "500 Connection Done",
			body: gin.H{
				"name":        strings.ToLower(ingredient.Name),
				"defaultUnit": ingredient.DefaultUnit,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.CreateIngredientParams{
					Name:        strings.ToLower(ingredient.Name),
					DefaultUnit: ingredient.DefaultUnit,
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					CreateIngredient(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Ingredient{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			url := "/ingredients/add"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))

			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestDeleteIngredientAPI(t *testing.T) {
	admin, _ := randomAdmin(t)
	ingredient := randomIngredient(t)

	testCases := []struct {
		name          string
		uri           int32
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker)
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			uri:  ingredient.ID,
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
					DeleteIngredient(gomock.Any(), gomock.Eq(ingredient.ID)).
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
					DeleteIngredient(gomock.Any(), gomock.Eq(-2)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "404 Not Found",
			uri:  ingredient.ID,
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
					DeleteIngredient(gomock.Any(), gomock.Eq(ingredient.ID)).
					Times(1).
					Return(sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "500 Internal Server Error",
			uri:  ingredient.ID,
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
					DeleteIngredient(gomock.Any(), gomock.Eq(ingredient.ID)).
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

			url := fmt.Sprintf("/ingredients/delete/%v", tc.uri)
			request, err := http.NewRequest(http.MethodDelete, url, nil)

			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetIngredientAPI(t *testing.T) {
	admin, _ := randomUser(t)
	ingredient := randomIngredient(t)

	testCases := []struct {
		name          string
		uri           int32
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker)
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			uri:  ingredient.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetIngredient(gomock.Any(), gomock.Eq(ingredient.ID)).
					Times(1).
					Return(ingredient, nil)
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
					GetIngredient(gomock.Any(), gomock.Eq(-2)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "404 Not Found",
			uri:  ingredient.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetIngredient(gomock.Any(), gomock.Eq(ingredient.ID)).
					Times(1).
					Return(db.Ingredient{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "500 Internal Server Error",
			uri:  ingredient.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetIngredient(gomock.Any(), gomock.Eq(ingredient.ID)).
					Times(1).
					Return(db.Ingredient{}, sql.ErrConnDone)
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

			url := fmt.Sprintf("/ingredients/%v", tc.uri)
			request, err := http.NewRequest(http.MethodGet, url, nil)

			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListIngredientsAPI(t *testing.T) {
	ingredients := []db.Ingredient{
		randomIngredient(t),
		randomIngredient(t),
		randomIngredient(t),
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
					ListIngredients(gomock.Any()).
					Times(1).
					Return(ingredients, nil)
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
					ListIngredients(gomock.Any()).
					Times(1).
					Return([]db.Ingredient{}, sql.ErrConnDone)
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

			url := "/ingredients/all"
			request, err := http.NewRequest(http.MethodGet, url, nil)

			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestSearchIngredientsAPI(t *testing.T) {
	ingredient := randomIngredient(t)
	ingredients := []db.SearchIngredientsRow{
		{
			ID: ingredient.ID,
			Name: ingredient.Name,
			DefaultUnit: ingredient.DefaultUnit,
		},
	}
	user, _ := randomUser(t)

	testCases := []struct {
		name          string
		query         string
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker)
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: "daging",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					SearchIngredients(gomock.Any(), gomock.Eq("%daging%")).
					Times(1).
					Return(ingredients, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "OK Empty Row",
			query: "daging",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					SearchIngredients(gomock.Any(), gomock.Eq("%daging%")).
					Times(1).
					Return([]db.SearchIngredientsRow{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "500 Internal Server Error",
			query: "daging",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					SearchIngredients(gomock.Any(), gomock.Eq("%daging%")).
					Times(1).
					Return([]db.SearchIngredientsRow{}, sql.ErrConnDone)
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

			url := fmt.Sprintf("/ingredients?name=%s", tc.query)
			request, err := http.NewRequest(http.MethodGet, url, nil)

			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateIngredientAPI(t *testing.T) {
	admin, _ := randomAdmin(t)
	ingredient := randomIngredient(t)

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
			uri: ingredient.ID,
			body: gin.H{
				"id":       ingredient.ID,
				"name": "new ingredient",
				"defaultUnit": gin.H{
					"Int32": 3,
					"Valid": true,
				},
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.UpdateIngredientParams{
					ID: ingredient.ID,
					Name: "new ingredient",
					DefaultUnit: sql.NullInt32{
						Int32: 3,
						Valid: true,
					},
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					UpdateIngredient(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Ingredient{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "400 Mismatched ID",
			uri: ingredient.ID,
			body: gin.H{
				"id":       ingredient.ID + 1,
				"name": "new ingredient",
				"defaultUnit": gin.H{
					"Int32": 3,
					"Valid": true,
				},
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.UpdateIngredientParams{
					ID: ingredient.ID,
					Name: "new ingredient",
					DefaultUnit: sql.NullInt32{
						Int32: 3,
						Valid: true,
					},
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					UpdateIngredient(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "400 Invalid UUID",
			uri: -ingredient.ID,
			body: gin.H{
				"id":       -ingredient.ID,
				"name": "new ingredient",
				"defaultUnit": gin.H{
					"Int32": 3,
					"Valid": true,
				},
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.UpdateIngredientParams{
					ID: -ingredient.ID,
					Name: "new ingredient",
					DefaultUnit: sql.NullInt32{
						Int32: 3,
						Valid: true,
					},
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					UpdateIngredient(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "404 Not Found",
			uri: ingredient.ID,
			body: gin.H{
				"id":       ingredient.ID,
				"name": "new ingredient",
				"defaultUnit": gin.H{
					"Int32": 3,
					"Valid": true,
				},
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.UpdateIngredientParams{
					ID: ingredient.ID,
					Name: "new ingredient",
					DefaultUnit: sql.NullInt32{
						Int32: 3,
						Valid: true,
					},
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					UpdateIngredient(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Ingredient{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "500 Internal Server Error",
			uri: ingredient.ID,
			body: gin.H{
				"id":       ingredient.ID,
				"name": "new ingredient",
				"defaultUnit": gin.H{
					"Int32": 3,
					"Valid": true,
				},
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.UpdateIngredientParams{
					ID: ingredient.ID,
					Name: "new ingredient",
					DefaultUnit: sql.NullInt32{
						Int32: 3,
						Valid: true,
					},
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					UpdateIngredient(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Ingredient{}, sql.ErrConnDone)
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

			url := fmt.Sprintf("/ingredients/update/%v", tc.uri)
			request, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func randomIngredient(t *testing.T) db.Ingredient {
	return db.Ingredient{
		ID:   int32(util.RandomInt(1, 300)),
		Name: util.RandomIngredient(),
		DefaultUnit: sql.NullInt32{
			Int32: int32(util.RandomInt(1, 300)),
			Valid: true,
		},
	}
}
