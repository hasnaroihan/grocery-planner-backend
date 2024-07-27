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
	"github.com/google/uuid"
	"github.com/hasnaroihan/grocery-planner/auth"
	dbmock "github.com/hasnaroihan/grocery-planner/db/mock"
	db "github.com/hasnaroihan/grocery-planner/db/sqlc"
	"github.com/hasnaroihan/grocery-planner/util"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

func TestGenerateGroceriesAPI(t *testing.T) {
	schedule := randomSchedule(uuid.NullUUID{})
	var scheduleRecipe []db.ScheduleRecipePortion
	for i := range schedule.Recipes {
		scheduleRecipe = append(scheduleRecipe, db.ScheduleRecipePortion{
			RecipeID: schedule.Recipes[i].RecipeID,
			Portion:  schedule.Recipes[i].Portion,
		})
	}

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"author":  uuid.NullUUID{},
				"recipes": scheduleRecipe,
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.GenerateGroceriesParam{
					Author:  uuid.NullUUID{},
					Recipes: scheduleRecipe,
				}
				storage.EXPECT().
					GenerateGroceries(gomock.Any(), arg).
					Times(1).
					Return(schedule, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "400 Empty Schedule",
			body: gin.H{
				"author":  uuid.NullUUID{},
				"recipes": db.ScheduleRecipePortion{},
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.GenerateGroceriesParam{
					Author:  uuid.NullUUID{},
					Recipes: scheduleRecipe,
				}
				storage.EXPECT().
					GenerateGroceries(gomock.Any(), arg).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "500 Internal Server Error",
			body: gin.H{
				"author":  uuid.NullUUID{},
				"recipes": scheduleRecipe,
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.GenerateGroceriesParam{
					Author:  uuid.NullUUID{},
					Recipes: scheduleRecipe,
				}
				storage.EXPECT().
					GenerateGroceries(gomock.Any(), arg).
					Times(1).
					Return(db.GenerateGroceriesResult{}, sql.ErrConnDone)
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

			url := "/groceries"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListSchedulesAPI(t *testing.T) {
	schedules := []db.Schedule{}
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
				arg := db.ListSchedulesParams{
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
					ListSchedules(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(schedules, nil)
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
				arg := db.ListSchedulesParams{
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
					ListSchedules(gomock.Any(), gomock.Eq(arg)).
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
				arg := db.ListSchedulesParams{
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
					ListSchedules(gomock.Any(), gomock.Eq(arg)).
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

			url := fmt.Sprintf("/schedule/all?%s", tc.query)
			request, err := http.NewRequest(http.MethodGet, url, nil)

			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListSchedulesUserAPI(t *testing.T) {
	user, _ := randomUser(t)
	admin, _ := randomAdmin(t)
	schedules := []db.Schedule{
		randomSchedule(uuid.NullUUID{UUID: user.ID, Valid: true}).Schedule,
		randomSchedule(uuid.NullUUID{UUID: user.ID, Valid: true}).Schedule,
		randomSchedule(uuid.NullUUID{UUID: user.ID, Valid: true}).Schedule,
	}

	testCases := []struct {
		name          string
		query         string
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker)
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK User",
			query: fmt.Sprintf("author=%s&pageSize=2&pageNum=1", user.ID.String()),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.ListSchedulesUserParams{
					Author: uuid.NullUUID{UUID: user.ID, Valid: true},
					Limit:  2,
					Offset: 0,
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role: "common",
					}, nil)
				storage.EXPECT().
					ListSchedulesUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(schedules, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:  "OK Admin",
			query: fmt.Sprintf("author=%s&pageSize=2&pageNum=1", user.ID.String()),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.ListSchedulesUserParams{
					Author: uuid.NullUUID{UUID: user.ID, Valid: true},
					Limit:  2,
					Offset: 0,
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role: "admin",
					}, nil)
				storage.EXPECT().
					ListSchedulesUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(schedules, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:  "403 Forbidden",
			query: fmt.Sprintf("author=%s&pageSize=2&pageNum=1", user.ID.String()),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				id, _ := uuid.NewRandom()
				addAuthorization(t, req, tokenMaker, authBearerType, id, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.ListSchedulesUserParams{
					Author: uuid.NullUUID{UUID: user.ID, Valid: true},
					Limit:  2,
					Offset: 0,
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.GetPermissionRow{
						Role: "common",
					}, nil)
				storage.EXPECT().
					ListSchedulesUser(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name:  "400 Bad Request",
			query: "pageSize=2",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.ListSchedulesUserParams{
					Author: uuid.NullUUID{UUID: user.ID, Valid: true},
					Limit:  2,
					Offset: 0,
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(user.ID)).
					Times(0)
				storage.EXPECT().
					ListSchedulesUser(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:  "500 Internal Server Error",
			query: fmt.Sprintf("author=%s&pageSize=2&pageNum=1", user.ID.String()),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.ListSchedulesUserParams{
					Author: uuid.NullUUID{UUID: user.ID, Valid: true},
					Limit:  2,
					Offset: 0,
				}
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role: "admin",
					}, nil)
				storage.EXPECT().
					ListSchedulesUser(gomock.Any(), gomock.Eq(arg)).
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

			url := fmt.Sprintf("/schedule/list?%s", tc.query)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestDeleteScheduleAPI(t *testing.T) {
	user, _ := randomUser(t)
	admin, _ := randomAdmin(t)
	schedule := randomSchedule(uuid.NullUUID{UUID: user.ID, Valid: true})

	testCases := []struct {
		name          string
		uri           int64
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker)
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK User",
			uri:  schedule.Schedule.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
					Times(1).
					Return(schedule.Schedule, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "common",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					DeleteSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "OK Admin",
			uri:  schedule.Schedule.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
					Times(1).
					Return(schedule.Schedule, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					DeleteSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "403 Forbidden",
			uri:  schedule.Schedule.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				id, err := uuid.NewRandom()
				require.NoError(t, err)
				addAuthorization(t, req, tokenMaker, authBearerType, id, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
					Times(1).
					Return(schedule.Schedule, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Not(user.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "common",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					DeleteSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
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
					GetSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
					Times(0)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(user.ID)).
					Times(0)
				storage.EXPECT().
					DeleteSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "404 Get Not Found",
			uri:  schedule.Schedule.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
					Times(1).
					Return(db.Schedule{}, sql.ErrNoRows)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Not(user.ID)).
					Times(0)
				storage.EXPECT().
					DeleteSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "500 Get Internal Server Error",
			uri:  schedule.Schedule.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
					Times(1).
					Return(db.Schedule{}, sql.ErrConnDone)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Not(user.ID)).
					Times(0)
				storage.EXPECT().
					DeleteSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "404 Delete Not Found",
			uri:  schedule.Schedule.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
					Times(1).
					Return(schedule.Schedule, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role: "common",
					}, nil)
				storage.EXPECT().
					DeleteSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
					Times(1).
					Return(sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "500 Delete Internal Server Error",
			uri:  schedule.Schedule.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				storage.EXPECT().
					GetSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
					Times(1).
					Return(schedule.Schedule, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role: "common",
					}, nil)
				storage.EXPECT().
					DeleteSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
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

			url := fmt.Sprintf("/schedule/delete/%d", tc.uri)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestDeleteScheduleRecipeAPI(t *testing.T) {
	user, _ := randomUser(t)
	admin, _ := randomAdmin(t)
	schedule := randomSchedule(uuid.NullUUID{UUID: user.ID, Valid: true})

	testCases := []struct {
		name          string
		query         string
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker)
		buildStubs    func(storage *dbmock.MockStorage)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK User",
			query: fmt.Sprintf("scheduleID=%d&recipeID=%d",
				schedule.Schedule.ID,
				schedule.Recipes[0].RecipeID),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.DeleteScheduleRecipeParams{
					ScheduleID: schedule.Schedule.ID,
					RecipeID:   schedule.Recipes[0].RecipeID,
				}
				storage.EXPECT().
					GetSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
					Times(1).
					Return(schedule.Schedule, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "common",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					DeleteScheduleRecipe(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "OK Admin",
			query: fmt.Sprintf("scheduleID=%d&recipeID=%d",
				schedule.Schedule.ID,
				schedule.Recipes[0].RecipeID),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, admin.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.DeleteScheduleRecipeParams{
					ScheduleID: schedule.Schedule.ID,
					RecipeID:   schedule.Recipes[0].RecipeID,
				}
				storage.EXPECT().
					GetSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
					Times(1).
					Return(schedule.Schedule, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "admin",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					DeleteScheduleRecipe(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "403 Forbidden",
			query: fmt.Sprintf("scheduleID=%d&recipeID=%d",
				schedule.Schedule.ID,
				schedule.Recipes[0].RecipeID),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				id, err := uuid.NewRandom()
				require.NoError(t, err)
				addAuthorization(t, req, tokenMaker, authBearerType, id, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.DeleteScheduleRecipeParams{
					ScheduleID: schedule.Schedule.ID,
					RecipeID:   schedule.Recipes[0].RecipeID,
				}
				storage.EXPECT().
					GetSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
					Times(1).
					Return(schedule.Schedule, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Not(user.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role:       "common",
						VerifiedAt: sql.NullTime{},
					}, nil)
				storage.EXPECT().
					DeleteScheduleRecipe(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name:  "400 Invalid ID",
			query: fmt.Sprintf("recipeID=%d&scheduleID=%d", -1, -1),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.DeleteScheduleRecipeParams{
					ScheduleID: schedule.Schedule.ID,
					RecipeID:   schedule.Recipes[0].RecipeID,
				}
				storage.EXPECT().
					GetSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
					Times(0)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(user.ID)).
					Times(0)
				storage.EXPECT().
					DeleteScheduleRecipe(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "404 Get Not Found",
			query: fmt.Sprintf("scheduleID=%d&recipeID=%d",
				schedule.Schedule.ID,
				schedule.Recipes[0].RecipeID),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.DeleteScheduleRecipeParams{
					ScheduleID: schedule.Schedule.ID,
					RecipeID:   schedule.Recipes[0].RecipeID,
				}
				storage.EXPECT().
					GetSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
					Times(1).
					Return(db.Schedule{}, sql.ErrNoRows)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Not(user.ID)).
					Times(0)
				storage.EXPECT().
					DeleteScheduleRecipe(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "500 Get Internal Server Error",
			query: fmt.Sprintf("scheduleID=%d&recipeID=%d",
				schedule.Schedule.ID,
				schedule.Recipes[0].RecipeID),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.DeleteScheduleRecipeParams{
					ScheduleID: schedule.Schedule.ID,
					RecipeID:   schedule.Recipes[0].RecipeID,
				}
				storage.EXPECT().
					GetSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
					Times(1).
					Return(db.Schedule{}, sql.ErrConnDone)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Not(user.ID)).
					Times(0)
				storage.EXPECT().
					DeleteScheduleRecipe(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "404 Delete Not Found",
			query: fmt.Sprintf("scheduleID=%d&recipeID=%d",
				schedule.Schedule.ID,
				schedule.Recipes[0].RecipeID),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.DeleteScheduleRecipeParams{
					ScheduleID: schedule.Schedule.ID,
					RecipeID:   schedule.Recipes[0].RecipeID,
				}
				storage.EXPECT().
					GetSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
					Times(1).
					Return(schedule.Schedule, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role: "common",
					}, nil)
				storage.EXPECT().
					DeleteScheduleRecipe(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "500 Delete Internal Server Error",
			query: fmt.Sprintf("scheduleID=%d&recipeID=%d",
				schedule.Schedule.ID,
				schedule.Recipes[0].RecipeID),
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker auth.TokenMaker) {
				addAuthorization(t, req, tokenMaker, authBearerType, user.ID, time.Minute)
			},
			buildStubs: func(storage *dbmock.MockStorage) {
				arg := db.DeleteScheduleRecipeParams{
					ScheduleID: schedule.Schedule.ID,
					RecipeID:   schedule.Recipes[0].RecipeID,
				}
				storage.EXPECT().
					GetSchedule(gomock.Any(), gomock.Eq(schedule.Schedule.ID)).
					Times(1).
					Return(schedule.Schedule, nil)
				storage.EXPECT().
					GetPermission(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(db.GetPermissionRow{
						Role: "common",
					}, nil)
				storage.EXPECT().
					DeleteScheduleRecipe(gomock.Any(), gomock.Eq(arg)).
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

			url := fmt.Sprintf("/schedule/delete?%s", tc.query)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func randomSchedule(Author uuid.NullUUID) db.GenerateGroceriesResult {
	scheduleId := util.RandomInt(1, 100)
	return db.GenerateGroceriesResult{
		Schedule: db.Schedule{
			ID:        scheduleId,
			Author:    Author,
			CreatedAt: time.Now(),
		},
		Recipes: []db.GetScheduleRecipeRow{
			{
				ScheduleID: scheduleId,
				RecipeID:   util.RandomInt(1, 100),
				Name:       util.RandomString(10),
				Portion:    int32(util.RandomInt(1, 5)),
			},
			{
				ScheduleID: scheduleId,
				RecipeID:   util.RandomInt(1, 100),
				Name:       util.RandomString(10),
				Portion:    int32(util.RandomInt(1, 5)),
			},
		},
		Groceries: []db.ListGroceriesRow{
			{
				ID:   int32(util.RandomInt(1, 100)),
				Name: util.RandomString(10),
			},
			{
				ID:   int32(util.RandomInt(1, 100)),
				Name: util.RandomString(10),
			},
			{
				ID:   int32(util.RandomInt(1, 100)),
				Name: util.RandomString(10),
			},
			{
				ID:   int32(util.RandomInt(1, 100)),
				Name: util.RandomString(10),
			},
		},
	}
}
