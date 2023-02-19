package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/hasnaroihan/grocery-planner/db/sqlc"
)

type createIngredientRequest struct {
	Name        string        `json:"name" binding:"required"`
	DefaultUnit sql.NullInt32 `json:"defaultUnit"`
}

func (server *Server) createIngredient(ctx *gin.Context) {
	var req createIngredientRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))

		return
	}

	arg := db.CreateIngredientParams{
		Name: req.Name,
		DefaultUnit: sql.NullInt32{
			Int32: req.DefaultUnit.Int32,
			Valid: req.DefaultUnit.Valid,
		},
	}

	ingredient, err := server.storage.CreateIngredient(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, ingredient)
}

type deleteIngredientRequest struct {
	ID int32	`uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteIngredient(ctx *gin.Context) {
	var req deleteIngredientRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))

		return
	}

	err := server.storage.DeleteIngredient(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, nil)
}