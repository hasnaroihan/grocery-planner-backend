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
	"github.com/google/uuid"
	"github.com/hasnaroihan/grocery-planner/auth"
	dbmock "github.com/hasnaroihan/grocery-planner/db/mock"
	db "github.com/hasnaroihan/grocery-planner/db/sqlc"
	"github.com/hasnaroihan/grocery-planner/util"
	"github.com/stretchr/testify/require"
)

func TestNewRecipeAPI(t *testing.T) {
	user, _ := randomUser(t)
	recipe := randomRecipe(user.ID)

	var ingredients []db.ListIngredientParam
	for i := range recipe.Ingredients {
		ingredients = append(ingredients, db.ListIngredientParam{
			ID: sql.NullInt32{
				Int32: recipe.Ingredients[i].IngredientID,
				Valid: true,
			},
			Name:   recipe.Ingredients[i].Name,
			Amount: recipe.Ingredients[i].Amount,
			UnitID: recipe.Ingredients[i].UnitID,
		})
	}

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker)
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"name":        recipe.Recipe.Name,
				"portion":     recipe.Recipe.Portion,
				"steps":       recipe.Recipe.Steps,
				"ingredients": ingredients,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.NewRecipeParams{
					Name:            recipe.Recipe.Name,
					Author:          user.ID,
					Portion:         recipe.Recipe.Portion,
					Steps:           recipe.Recipe.Steps,
					ListIngredients: ingredients,
				}
				storage.EXPECT().
					NewRecipeTx(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(recipe, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "400 Invalid Portion",
			body: gin.H{
				"name":        recipe.Recipe.Name,
				"portion":     "2",
				"steps":       recipe.Recipe.Steps,
				"ingredients": ingredients,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.NewRecipeParams{
					Name:            recipe.Recipe.Name,
					Author:          user.ID,
					Portion:         recipe.Recipe.Portion,
					Steps:           recipe.Recipe.Steps,
					ListIngredients: ingredients,
				}
				storage.EXPECT().
					NewRecipeTx(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "400 Invalid Steps",
			body: gin.H{
				"name":        recipe.Recipe.Name,
				"portion":     recipe.Recipe.Portion,
				"steps":       "ä",
				"ingredients": ingredients,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.NewRecipeParams{
					Name:            recipe.Recipe.Name,
					Author:          user.ID,
					Portion:         recipe.Recipe.Portion,
					Steps:           recipe.Recipe.Steps,
					ListIngredients: ingredients,
				}
				storage.EXPECT().
					NewRecipeTx(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "400 Empty Ingredients",
			body: gin.H{
				"name":        recipe.Recipe.Name,
				"portion":     recipe.Recipe.Portion,
				"steps":       recipe.Recipe.Steps,
				"ingredients": gin.H{},
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.NewRecipeParams{
					Name:            recipe.Recipe.Name,
					Author:          user.ID,
					Portion:         recipe.Recipe.Portion,
					Steps:           recipe.Recipe.Steps,
					ListIngredients: ingredients,
				}
				storage.EXPECT().
					NewRecipeTx(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "500 Internal Server Error",
			body: gin.H{
				"name":        recipe.Recipe.Name,
				"portion":     recipe.Recipe.Portion,
				"steps":       recipe.Recipe.Steps,
				"ingredients": ingredients,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.NewRecipeParams{
					Name:            recipe.Recipe.Name,
					Author:          user.ID,
					Portion:         recipe.Recipe.Portion,
					Steps:           recipe.Recipe.Steps,
					ListIngredients: ingredients,
				}
				storage.EXPECT().
					NewRecipeTx(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.RecipeResult{}, sql.ErrConnDone)
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

			url := "/recipe/add"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestDeleteRecipeAPI(t *testing.T) {
	user, _ := randomUser(t)
	admin, _ := randomAdmin(t)
	recipe := randomRecipe(user.ID)

	testCases := []struct {
		name          string
		uri           int64
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker)
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK User",
			uri:  recipe.Recipe.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(recipe.Recipe, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "common",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					DeleteRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "OK Admin",
			uri:  recipe.Recipe.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(recipe.Recipe, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					DeleteRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "403 Forbidden",
			uri:  recipe.Recipe.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				id, err := uuid.NewRandom()
				require.NoError(t, err)
				addAuthorization(t, req, tokenMaker, authBearerType, id, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(recipe.Recipe, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Not(user.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "common",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					DeleteRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "400 Invalid ID",
			uri:  -1,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(0)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(user.ID)).
					Times(0)
				storage.EXPECT().
					DeleteRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "404 Get Not Found",
			uri:  recipe.Recipe.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(db.Recipe{}, sql.ErrNoRows)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Not(user.ID)).
					Times(0)
				storage.EXPECT().
					DeleteRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "500 Get Internal Server Error",
			uri:  recipe.Recipe.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(db.Recipe{}, sql.ErrConnDone)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Not(user.ID)).
					Times(0)
				storage.EXPECT().
					DeleteRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "404 Delete Not Found",
			uri:  recipe.Recipe.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(recipe.Recipe, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role: "common",
					}, nil)
				storage.EXPECT().
					DeleteRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "500 Delete Internal Server Error",
			uri:  recipe.Recipe.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(recipe.Recipe, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role: "common",
					}, nil)
				storage.EXPECT().
					DeleteRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
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

			url := fmt.Sprintf("/recipe/delete/%d", tc.uri)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestDeleteRecipeIngredientAPI(t *testing.T) {
	user, _ := randomUser(t)
	admin, _ := randomAdmin(t)
	recipe := randomRecipe(user.ID)

	testCases := []struct {
		name          string
		query         string
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker)
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK User",
			query: fmt.Sprintf("recipeID=%d&ingredientID=%d",
				recipe.Recipe.ID,
				recipe.Ingredients[0].IngredientID),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.DeleteRecipeIngredientParams{
					RecipeID:     recipe.Recipe.ID,
					IngredientID: recipe.Ingredients[0].IngredientID,
				}
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(recipe.Recipe, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "common",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					DeleteRecipeIngredient(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "OK Admin",
			query: fmt.Sprintf("recipeID=%d&ingredientID=%d",
				recipe.Recipe.ID,
				recipe.Ingredients[0].IngredientID),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.DeleteRecipeIngredientParams{
					RecipeID:     recipe.Recipe.ID,
					IngredientID: recipe.Ingredients[0].IngredientID,
				}
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(recipe.Recipe, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					DeleteRecipeIngredient(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "403 Forbidden",
			query: fmt.Sprintf("recipeID=%d&ingredientID=%d",
				recipe.Recipe.ID,
				recipe.Ingredients[0].IngredientID),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				id, err := uuid.NewRandom()
				require.NoError(t, err)
				addAuthorization(t, req, tokenMaker, authBearerType, id, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.DeleteRecipeIngredientParams{
					RecipeID:     recipe.Recipe.ID,
					IngredientID: recipe.Ingredients[0].IngredientID,
				}
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(recipe.Recipe, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Not(user.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "common",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					DeleteRecipeIngredient(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name:  "400 Invalid ID",
			query: fmt.Sprintf("recipeID=%d&ingredientID=%d", -1, -1),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.DeleteRecipeIngredientParams{
					RecipeID:     recipe.Recipe.ID,
					IngredientID: recipe.Ingredients[0].IngredientID,
				}
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(0)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(user.ID)).
					Times(0)
				storage.EXPECT().
					DeleteRecipeIngredient(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "404 Get Not Found",
			query: fmt.Sprintf("recipeID=%d&ingredientID=%d",
				recipe.Recipe.ID,
				recipe.Ingredients[0].IngredientID),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.DeleteRecipeIngredientParams{
					RecipeID:     recipe.Recipe.ID,
					IngredientID: recipe.Ingredients[0].IngredientID,
				}
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(db.Recipe{}, sql.ErrNoRows)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Not(user.ID)).
					Times(0)
				storage.EXPECT().
					DeleteRecipeIngredient(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "500 Get Internal Server Error",
			query: fmt.Sprintf("recipeID=%d&ingredientID=%d",
				recipe.Recipe.ID,
				recipe.Ingredients[0].IngredientID),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.DeleteRecipeIngredientParams{
					RecipeID:     recipe.Recipe.ID,
					IngredientID: recipe.Ingredients[0].IngredientID,
				}
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(db.Recipe{}, sql.ErrConnDone)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Not(user.ID)).
					Times(0)
				storage.EXPECT().
					DeleteRecipeIngredient(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "404 Delete Not Found",
			query: fmt.Sprintf("recipeID=%d&ingredientID=%d",
				recipe.Recipe.ID,
				recipe.Ingredients[0].IngredientID),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.DeleteRecipeIngredientParams{
					RecipeID:     recipe.Recipe.ID,
					IngredientID: recipe.Ingredients[0].IngredientID,
				}
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(recipe.Recipe, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role: "common",
					}, nil)
				storage.EXPECT().
					DeleteRecipeIngredient(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "500 Delete Internal Server Error",
			query: fmt.Sprintf("recipeID=%d&ingredientID=%d",
				recipe.Recipe.ID,
				recipe.Ingredients[0].IngredientID),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.DeleteRecipeIngredientParams{
					RecipeID:     recipe.Recipe.ID,
					IngredientID: recipe.Ingredients[0].IngredientID,
				}
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(recipe.Recipe, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role: "common",
					}, nil)
				storage.EXPECT().
					DeleteRecipeIngredient(gomock.Any(), gomock.Eq(arg)).
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

			url := fmt.Sprintf("/recipe/delete?%s", tc.query)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetRecipeAPI(t *testing.T) {
	user, _ := randomUser(t)
	recipe := randomRecipe(user.ID)

	testCases := []struct {
		name          string
		uri           int64
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			uri: recipe.Recipe.ID,
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetRecipeTx(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(recipe, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "400 Bad request",
			uri: -2,
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetRecipeTx(gomock.Any(), gomock.Eq(-2)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "404 Not Found",
			uri: recipe.Recipe.ID,
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetRecipeTx(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(db.RecipeResult{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "500 Internal Server Error",
			uri: recipe.Recipe.ID,
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetRecipeTx(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(db.RecipeResult{}, sql.ErrTxDone)
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

			url := fmt.Sprintf("/recipe/%d", tc.uri)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListRecipesAPI(t *testing.T) {
	user,_ := randomUser(t)
	recipes := []db.Recipe {
		randomRecipe(user.ID).Recipe,
		randomRecipe(user.ID).Recipe,
		randomRecipe(user.ID).Recipe,
	}

	testCases := []struct {
		name          string
		query         string
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: "pageSize=2&pageNum=1",
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.ListRecipesParams{
					Limit:  2,
					Offset: 0,
				}
				storage.EXPECT().
					ListRecipes(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(recipes, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:  "400 Bad Request",
			query: "pageSize=2",
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.ListRecipesParams{
					Limit:  2,
					Offset: 0,
				}
				storage.EXPECT().
					ListRecipes(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:  "500 Internal Server Error",
			query: "pageSize=2&pageNum=1",
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.ListRecipesParams{
					Limit:  2,
					Offset: 0,
				}
				storage.EXPECT().
					ListRecipes(gomock.Any(), gomock.Eq(arg)).
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

			url := fmt.Sprintf("/recipe/all?%s", tc.query)
			request, err := http.NewRequest(http.MethodGet, url, nil)

			require.NoError(t, err)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListRecipesUserAPI(t *testing.T) {
	user,_ := randomUser(t)
	recipes := []db.Recipe {
		randomRecipe(user.ID).Recipe,
		randomRecipe(user.ID).Recipe,
		randomRecipe(user.ID).Recipe,
	}

	testCases := []struct {
		name          string
		query         string
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: fmt.Sprintf("author=%s&pageSize=2&pageNum=1", user.ID.String()),
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.ListRecipesUserParams{
					Author: user.ID,
					Limit:  2,
					Offset: 0,
				}
				storage.EXPECT().
					ListRecipesUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(recipes, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:  "400 Bad Request",
			query: "pageSize=2",
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.ListRecipesUserParams{
					Author: user.ID,
					Limit:  2,
					Offset: 0,
				}
				storage.EXPECT().
					ListRecipesUser(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:  "500 Internal Server Error",
			query: fmt.Sprintf("author=%s&pageSize=2&pageNum=1", user.ID.String()),
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.ListRecipesUserParams{
					Author: user.ID,
					Limit:  2,
					Offset: 0,
				}
				storage.EXPECT().
					ListRecipesUser(gomock.Any(), gomock.Eq(arg)).
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

			url := fmt.Sprintf("/recipe/list?%s", tc.query)
			request, err := http.NewRequest(http.MethodGet, url, nil)

			require.NoError(t, err)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestSearchRecipesAPI(t *testing.T) {
	user,_ := randomUser(t)
	rec := randomRecipe(user.ID)
	recipes := []db.SearchRecipeRow {
		{
			ID: rec.Recipe.ID,
			Name: rec.Recipe.Name,
			Author: rec.Recipe.Author,
			ModifiedAt: rec.Recipe.ModifiedAt,
		},
	}

	testCases := []struct {
		name          string
		query         string
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: fmt.Sprintf("name=%s&pageSize=2&pageNum=1", strings.ToLower(recipes[0].Name)),
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.SearchRecipeParams{
					Name:	fmt.Sprintf("%%%s%%", strings.ToLower(recipes[0].Name)),
					Limit:  2,
					Offset: 0,
				}
				storage.EXPECT().
					SearchRecipe(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(recipes, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:  "400 Bad Request",
			query: fmt.Sprintf("name=%s&pageSize=2&pageNum=1", recipes[0].Name),
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.SearchRecipeParams{
					Name:	fmt.Sprintf("%%%s%%", recipes[0].Name),
					Limit:  2,
					Offset: 0,
				}
				storage.EXPECT().
					SearchRecipe(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:  "500 Internal Server Error",
			query: fmt.Sprintf("name=%s&pageSize=2&pageNum=1", strings.ToLower(recipes[0].Name)),
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.SearchRecipeParams{
					Name:	fmt.Sprintf("%%%s%%", strings.ToLower(recipes[0].Name)),
					Limit:  2,
					Offset: 0,
				}
				storage.EXPECT().
					SearchRecipe(gomock.Any(), gomock.Eq(arg)).
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

			url := fmt.Sprintf("/recipe?%s", tc.query)
			request, err := http.NewRequest(http.MethodGet, url, nil)

			require.NoError(t, err)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateRecipeAPI(t *testing.T) {
	user,_ := randomUser(t)
	admin,_ := randomAdmin(t)
	recipe := randomRecipe(user.ID)
	var ingredients []db.ListIngredientParam
	for i := range recipe.Ingredients {
		ingredients = append(ingredients, db.ListIngredientParam{
			ID: sql.NullInt32{
				Int32: recipe.Ingredients[i].IngredientID,
				Valid: true,
			},
			Name:   recipe.Ingredients[i].Name,
			Amount: recipe.Ingredients[i].Amount,
			UnitID: recipe.Ingredients[i].UnitID,
		})
	}

	testCases := []struct {
		name string
		uri int64
		body gin.H
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker)
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK User",
			uri: recipe.Recipe.ID,
			body: gin.H{
				"id": recipe.Recipe.ID,
				"name": "new recipe name",
				"portion": 5,
				"steps": gin.H {
					"String": "step 123",
					"Valid": true,
				},
				"ingredients": ingredients,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker){
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs   : func(storage *dbmock.MockStorage) {
				arg := db.TxUpdateRecipeParams {
					Recipe: db.UpdateRecipeParams{
						ID: recipe.Recipe.ID,
						Name: "new recipe name",
						Portion: 5,
						Steps: sql.NullString{
							String: "step 123",
							Valid: true,
						},
					},
					ListIngredients: ingredients,
				}
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(recipe.Recipe, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role: "common",
					}, nil)
				storage.EXPECT().
					UpdateRecipeTx(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(recipe, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder){
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "OK Admin",
			uri: recipe.Recipe.ID,
			body: gin.H{
				"id": recipe.Recipe.ID,
				"name": "new recipe name",
				"portion": 5,
				"steps": gin.H {
					"String": "step 123",
					"Valid": true,
				},
				"ingredients": ingredients,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker){
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs   : func(storage *dbmock.MockStorage) {
				arg := db.TxUpdateRecipeParams {
					Recipe: db.UpdateRecipeParams{
						ID: recipe.Recipe.ID,
						Name: "new recipe name",
						Portion: 5,
						Steps: sql.NullString{
							String: "step 123",
							Valid: true,
						},
					},
					ListIngredients: ingredients,
				}
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(recipe.Recipe, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role: "admin",
					}, nil)
				storage.EXPECT().
					UpdateRecipeTx(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(recipe, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder){
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "400 Mismatched ID",
			uri: util.RandomInt(1,100),
			body: gin.H{
				"id": recipe.Recipe.ID,
				"name": "new recipe name",
				"portion": 5,
				"steps": gin.H {
					"String": "step 123",
					"Valid": true,
				},
				"ingredients": ingredients,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker){
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs   : func(storage *dbmock.MockStorage) {
				arg := db.TxUpdateRecipeParams {
					Recipe: db.UpdateRecipeParams{
						ID: recipe.Recipe.ID,
						Name: "new recipe name",
						Portion: 5,
						Steps: sql.NullString{
							String: "step 123",
							Valid: true,
						},
					},
					ListIngredients: ingredients,
				}
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(0)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(0)
				storage.EXPECT().
					UpdateRecipeTx(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder){
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "400 Invalid URI",
			uri: -2,
			body: gin.H{
				"id": recipe.Recipe.ID,
				"name": "new recipe name",
				"portion": 5,
				"steps": gin.H {
					"String": "step 123",
					"Valid": true,
				},
				"ingredients": ingredients,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker){
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs   : func(storage *dbmock.MockStorage) {
				arg := db.TxUpdateRecipeParams {
					Recipe: db.UpdateRecipeParams{
						ID: recipe.Recipe.ID,
						Name: "new recipe name",
						Portion: 5,
						Steps: sql.NullString{
							String: "step 123",
							Valid: true,
						},
					},
					ListIngredients: ingredients,
				}
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(0)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(0)
				storage.EXPECT().
					UpdateRecipeTx(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder){
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "400 Invalid ID JSON",
			uri: recipe.Recipe.ID,
			body: gin.H{
				"id": -2,
				"name": "new recipe name",
				"portion": 5,
				"steps": gin.H {
					"String": "step 123",
					"Valid": true,
				},
				"ingredients": ingredients,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker){
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs   : func(storage *dbmock.MockStorage) {
				arg := db.TxUpdateRecipeParams {
					Recipe: db.UpdateRecipeParams{
						ID: recipe.Recipe.ID,
						Name: "new recipe name",
						Portion: 5,
						Steps: sql.NullString{
							String: "step 123",
							Valid: true,
						},
					},
					ListIngredients: ingredients,
				}
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(0)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(0)
				storage.EXPECT().
					UpdateRecipeTx(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder){
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "400 Invalid Name",
			uri: recipe.Recipe.ID,
			body: gin.H{
				"id": recipe.Recipe.ID,
				"name": "new recipe NAME",
				"portion": 5,
				"steps": gin.H {
					"String": "step 123",
					"Valid": true,
				},
				"ingredients": ingredients,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker){
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs   : func(storage *dbmock.MockStorage) {
				arg := db.TxUpdateRecipeParams {
					Recipe: db.UpdateRecipeParams{
						ID: recipe.Recipe.ID,
						Name: "new recipe name",
						Portion: 5,
						Steps: sql.NullString{
							String: "step 123",
							Valid: true,
						},
					},
					ListIngredients: ingredients,
				}
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(0)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(0)
				storage.EXPECT().
					UpdateRecipeTx(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder){
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "400 Invalid Portion",
			uri: recipe.Recipe.ID,
			body: gin.H{
				"id": recipe.Recipe.ID,
				"name": "new recipe name",
				"portion": 0,
				"steps": gin.H {
					"String": "step 123",
					"Valid": true,
				},
				"ingredients": ingredients,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker){
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs   : func(storage *dbmock.MockStorage) {
				arg := db.TxUpdateRecipeParams {
					Recipe: db.UpdateRecipeParams{
						ID: recipe.Recipe.ID,
						Name: "new recipe name",
						Portion: 5,
						Steps: sql.NullString{
							String: "step 123",
							Valid: true,
						},
					},
					ListIngredients: ingredients,
				}
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(0)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(0)
				storage.EXPECT().
					UpdateRecipeTx(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder){
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "400 Invalid Ingredients",
			uri: recipe.Recipe.ID,
			body: gin.H{
				"id": recipe.Recipe.ID,
				"name": "new recipe name",
				"portion": 5,
				"steps": gin.H {
					"String": "step 123",
					"Valid": true,
				},
				"ingredients": db.ListIngredientParam{},
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker){
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs   : func(storage *dbmock.MockStorage) {
				arg := db.TxUpdateRecipeParams {
					Recipe: db.UpdateRecipeParams{
						ID: recipe.Recipe.ID,
						Name: "new recipe name",
						Portion: 5,
						Steps: sql.NullString{
							String: "step 123",
							Valid: true,
						},
					},
					ListIngredients: ingredients,
				}
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(0)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(0)
				storage.EXPECT().
					UpdateRecipeTx(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder){
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "403 Forbidden",
			uri: recipe.Recipe.ID,
			body: gin.H{
				"id": recipe.Recipe.ID,
				"name": "new recipe name",
				"portion": 5,
				"steps": gin.H {
					"String": "step 123",
					"Valid": true,
				},
				"ingredients": ingredients,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker){
				id,err := uuid.NewRandom()
				require.NoError(t, err)
				addAuthorization(t, req, tokenMaker, authBearerType, id, time.Minute)
			},
			buildStubs   : func(storage *dbmock.MockStorage) {
				arg := db.TxUpdateRecipeParams {
					Recipe: db.UpdateRecipeParams{
						ID: recipe.Recipe.ID,
						Name: "new recipe name",
						Portion: 5,
						Steps: sql.NullString{
							String: "step 123",
							Valid: true,
						},
					},
					ListIngredients: ingredients,
				}
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(recipe.Recipe, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Not(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role: "common",
					}, nil)
				storage.EXPECT().
					UpdateRecipeTx(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder){
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "404 Get Not Found",
			uri: recipe.Recipe.ID,
			body: gin.H{
				"id": recipe.Recipe.ID,
				"name": "new recipe name",
				"portion": 5,
				"steps": gin.H {
					"String": "step 123",
					"Valid": true,
				},
				"ingredients": ingredients,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker){
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs   : func(storage *dbmock.MockStorage) {
				arg := db.TxUpdateRecipeParams {
					Recipe: db.UpdateRecipeParams{
						ID: recipe.Recipe.ID,
						Name: "new recipe name",
						Portion: 5,
						Steps: sql.NullString{
							String: "step 123",
							Valid: true,
						},
					},
					ListIngredients: ingredients,
				}
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(db.Recipe{}, sql.ErrNoRows)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(0)
				storage.EXPECT().
					UpdateRecipeTx(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder){
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "500 Get Internal Server Error",
			uri: recipe.Recipe.ID,
			body: gin.H{
				"id": recipe.Recipe.ID,
				"name": "new recipe name",
				"portion": 5,
				"steps": gin.H {
					"String": "step 123",
					"Valid": true,
				},
				"ingredients": ingredients,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker){
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs   : func(storage *dbmock.MockStorage) {
				arg := db.TxUpdateRecipeParams {
					Recipe: db.UpdateRecipeParams{
						ID: recipe.Recipe.ID,
						Name: "new recipe name",
						Portion: 5,
						Steps: sql.NullString{
							String: "step 123",
							Valid: true,
						},
					},
					ListIngredients: ingredients,
				}
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(db.Recipe{}, sql.ErrConnDone)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(0)
				storage.EXPECT().
					UpdateRecipeTx(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder){
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "500 Update Internal Server Error",
			uri: recipe.Recipe.ID,
			body: gin.H{
				"id": recipe.Recipe.ID,
				"name": "new recipe name",
				"portion": 5,
				"steps": gin.H {
					"String": "step 123",
					"Valid": true,
				},
				"ingredients": ingredients,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker){
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs   : func(storage *dbmock.MockStorage) {
				arg := db.TxUpdateRecipeParams {
					Recipe: db.UpdateRecipeParams{
						ID: recipe.Recipe.ID,
						Name: "new recipe name",
						Portion: 5,
						Steps: sql.NullString{
							String: "step 123",
							Valid: true,
						},
					},
					ListIngredients: ingredients,
				}
				storage.EXPECT().
					GetRecipe(gomock.Any(), gomock.Eq(recipe.Recipe.ID)).
					Times(1).
					Return(recipe.Recipe, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role: "common",
					}, nil)
				storage.EXPECT().
					UpdateRecipeTx(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.RecipeResult{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder){
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func (t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storage := dbmock.NewMockStorage(ctrl)
			tc.buildStubs(storage)

			server := newTestServer(t, storage)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/recipe/update/%d", tc.uri)
			request, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func randomRecipe(author uuid.UUID) db.RecipeResult {
	recipeId := util.RandomInt(1, 100)
	return db.RecipeResult{
		Recipe: db.Recipe{
			ID:      recipeId,
			Name:    util.RandomString(10),
			Author:  author,
			Portion: int32(util.RandomInt(1, 10)),
			Steps: sql.NullString{
				String: util.RandomString(100),
				Valid:  true,
			},
			CreatedAt:  time.Now().UTC().Add(time.Second),
			ModifiedAt: time.Now().UTC().Add(time.Second),
		},
		Ingredients: []db.GetRecipeIngredientsRow{
			{
				RecipeID:     recipeId,
				IngredientID: int32(util.RandomInt(1, 100)),
				Name:         util.RandomString(10),
				Amount:       float32(util.RandomInt(1, 500)),
				UnitID:       int32(util.RandomInt(1, 100)),
			},
			{
				RecipeID:     recipeId,
				IngredientID: int32(util.RandomInt(1, 100)),
				Name:         util.RandomString(10),
				Amount:       float32(util.RandomInt(1, 500)),
				UnitID:       int32(util.RandomInt(1, 100)),
			},
			{
				RecipeID:     recipeId,
				IngredientID: int32(util.RandomInt(1, 100)),
				Name:         util.RandomString(10),
				Amount:       float32(util.RandomInt(1, 500)),
				UnitID:       int32(util.RandomInt(1, 100)),
			},
		},
	}
}
