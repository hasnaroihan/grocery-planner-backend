package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/hasnaroihan/grocery-planner/db/sqlc"
)

type createUnitRequest struct {
	Name string `json:"name" binding:"required"`
}

func (server *Server) createUnit(ctx *gin.Context) {
	var req createUnitRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))

		return
	}

	unit, err := server.storage.CreateUnit(ctx, req.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, unit)
}

type deleteUnitRequest struct {
	ID int32	`uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteUnit(ctx *gin.Context) {
	var req deleteUnitRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))

		return
	}

	err := server.storage.DeleteUnit(ctx, req.ID)
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

type getUnitRequest struct {
	ID int32	`uri:"id" binding:"required,min=1"`
}

func (server *Server) getUnit(ctx *gin.Context) {
	var req getUnitRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	unit, err := server.storage.GetUnit(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, unit)
}

func (server *Server) listUnits(ctx *gin.Context) {
	units, err := server.storage.ListUnits(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, units)
}

type updateUnitUri struct {
	ID int32	`uri:"id" binding:"required,min=1"`
}

type updateUnitJSON struct {
	ID			int32		  `json:"id" binding:"required,min=1"`
	Name        string        `json:"name" binding:"required"`
	DefaultUnit sql.NullInt32 `json:"defaultUnit"`
}

func (server *Server) updateUnit(ctx *gin.Context) {
	var reqUri updateUnitUri
	var reqJSON updateUnitJSON

	if err := ctx.ShouldBindUri(&reqUri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindJSON(&reqJSON); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if reqUri.ID != reqJSON.ID {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("mismatched uri and body ingredient id")))
		return
	}

	arg := db.UpdateUnitParams{
		ID: reqUri.ID,
		Name: reqJSON.Name,
	}

	unit, err := server.storage.UpdateUnit(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, unit)
}