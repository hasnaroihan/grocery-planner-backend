package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hasnaroihan/grocery-planner/auth"
	db "github.com/hasnaroihan/grocery-planner/db/sqlc"
)

type generateGroceriesRequest struct {
	Author  uuid.NullUUID           `json:"author" binding:"required"`
	Recipes []db.ScheduleRecipePortion `json:"recipes" binding:"required,min=1"`
}

func (server *Server) generateGroceries(ctx *gin.Context) {
	var req generateGroceriesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.GenerateGroceriesParam{
		Author: req.Author,
		Recipes: req.Recipes,
	}

	groceries, err := server.storage.GenerateGroceries(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, groceries)
}

func (server *Server) listSchedules(ctx *gin.Context) {
	var req listPageRequest
	
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListSchedulesParams {
		Limit: req.PageSize,
		Offset: req.PageNum,
	}
	recipes, err := server.storage.ListSchedules(ctx,arg)
	if err != nil {
		if err != sql.ErrNoRows {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}

	ctx.JSON(http.StatusOK, recipes)
}

type listPageUserRequest struct {
	ID 			uuid.UUID	`form:"author" binding:"required,uuid"`
	PageSize	int32 		`form:"pageSize" binding:"required,number"`
	PageNum		int32 		`form:"pageNum" binding:"required,number"`
}

func (server *Server) listSchedulesUser(ctx *gin.Context) {
	var req listPageUserRequest
		
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// check permission
	authPayload := ctx.MustGet(authPayloadKey).(*auth.Payload)
	permit, err := server.storage.GetPermission(ctx, authPayload.Subject)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if req.ID != authPayload.Subject || permit.Role != "admin" {
		ctx.JSON(http.StatusForbidden, errorResponse(ErrAccessDenied))
		return
	}

	arg := db.ListSchedulesUserParams {
		Author: uuid.NullUUID{
			UUID: req.ID,
			Valid: true,
		},
		Limit: req.PageSize,
		Offset: req.PageNum,
	}
	recipes, err := server.storage.ListSchedulesUser(ctx,arg)
	if err != nil {
		if err != sql.ErrNoRows {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}

	ctx.JSON(http.StatusOK, recipes)
}

type deleteScheduleRequest struct {
	ID int64	`uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteSchedule(ctx *gin.Context) {
	var req deleteScheduleRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))

		return
	}

	schedule, err := server.storage.GetSchedule(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	
	// check permission
	authPayload := ctx.MustGet(authPayloadKey).(*auth.Payload)
	permit, err := server.storage.GetPermission(ctx, authPayload.Subject)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if schedule.Author.UUID != authPayload.Subject || permit.Role != "admin" {
		ctx.JSON(http.StatusForbidden, errorResponse(ErrAccessDenied))
		return
	}

	err = server.storage.DeleteSchedule(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

type deleteScheduleRecipeRequest struct {
	ScheduleID     int64 `form:"scheduleID" binding:"required,min=1"`
	RecipeID       int64 `form:"recipientID" binding:"required,min=1"`
}

func (server *Server) deleteScheduleRecipe(ctx *gin.Context) {
	var req deleteScheduleRecipeRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))

		return
	}

	schedule, err := server.storage.GetSchedule(ctx, req.ScheduleID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	
	// check permission
	authPayload := ctx.MustGet(authPayloadKey).(*auth.Payload)
	permit, err := server.storage.GetPermission(ctx, authPayload.Subject)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if schedule.Author.UUID != authPayload.Subject || permit.Role != "admin" {
		ctx.JSON(http.StatusForbidden, errorResponse(ErrAccessDenied))
		return
	}

	arg := db.DeleteScheduleRecipeParams {
		ScheduleID: req.ScheduleID,
		RecipeID: req.RecipeID,
	}

	err = server.storage.DeleteScheduleRecipe(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, nil)
}