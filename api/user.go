package api

import (
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/hasnaroihan/grocery-planner/auth"
	db "github.com/hasnaroihan/grocery-planner/db/sqlc"
	"github.com/hasnaroihan/grocery-planner/util"
	"github.com/lib/pq"
)

type registerUserRequest struct {
	Username string `json:"username" binding:"required,username"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type userResponse struct {
	ID         uuid.UUID    `json:"id"`
	Username   string       `json:"username"`
	Email      string       `json:"email"`
	CreatedAt  time.Time    `json:"createdAt"`
	Role       string       `json:"role"`
	VerifiedAt sql.NullTime `json:"verifiedAt"`
}

var usernameValidator validator.Func = func(fl validator.FieldLevel) bool {
	var regex,_ = regexp.Compile(`^[\w.-_]*$`)
	username := fl.Field().String()

	isMatch := regex.MatchString(username)

	return isMatch
}

func (server *Server) registerUser(ctx *gin.Context) {
	var req registerUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))

		return
	}

	hashPass, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}

	arg := db.CreateUserParams{
		Username: req.Username,
		Email: req.Email,
		Password: hashPass,
		Role: "common",
	}

	user, err := server.storage.CreateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code.Name() == "unique_violation" {
				ctx.JSON(http.StatusForbidden, errorResponse(err))
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := userResponse {
		ID: user.ID,
		Username: user.Username,
		Email: user.Email,
		CreatedAt: user.CreatedAt,
		VerifiedAt: user.VerifiedAt,
		Role: user.Role,
	}

	ctx.JSON(http.StatusOK, response)
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,username"`
	Password string `json:"password" binding:"required,min=8"`
}

type loginUserResponse struct {
	AccessToken string 			`json:"mice"`
	User 		userResponse	`json:"user"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.storage.GetLogin(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = util.ComparePassword(req.Password, user.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, err := server.tokenMaker.CreateToken(
		user.ID,
		server.tokenDuration,
		[]string{},
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	response := loginUserResponse{
		AccessToken: accessToken,
		User: userResponse{
			ID: user.ID,
			Username: user.Username,
			CreatedAt: user.CreatedAt,
			VerifiedAt: user.VerifiedAt,
			Role: user.Role,
		},
	}

	ctx.JSON(http.StatusOK, response)
}

type deleteUserRequest struct {
	ID uuid.UUID	`uri:"id" binding:"required,uuid4"`
}

func (server *Server) deleteUser(ctx *gin.Context) {
	var req deleteUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))

		return
	}

	err := server.storage.DeleteUser(ctx, req.ID)
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

type getUserRequest struct {
	ID         uuid.UUID    `json:"id" binding:"required,uuid"`
}

func (server *Server) getUser(ctx *gin.Context) {
	var req getUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Check permission
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

	user, err := server.storage.GetUser(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := userResponse {
		ID: user.ID,
		Username: user.Username,
		Email: user.Email,
		CreatedAt: user.CreatedAt,
		VerifiedAt: user.VerifiedAt,
		Role: user.Role,
	}

	ctx.JSON(http.StatusOK, response)
}

func (server *Server) listUsers(ctx *gin.Context) {
	var req listPageRequest
	
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListUsersParams {
		Limit: req.PageSize,
		Offset: req.PageNum,
	}
	users, err := server.storage.ListUsers(ctx,arg)
	if err != nil {
		if err != sql.ErrNoRows {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}

	ctx.JSON(http.StatusOK, users)
}

type updateUserUri struct {
	ID uuid.UUID	`uri:"id" binding:"required,uuid4"`
}

type updateUserJSON struct {
	ID         uuid.UUID    `json:"id" binding:"required,uuid4"`
	Username   string       `json:"username" binding:"required,username"`
	Email      string       `json:"email" binding:"required,email"`
}

func (server *Server) updateUser(ctx *gin.Context) {
	var reqUri updateUserUri
	var reqJSON updateUserJSON

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

	// Check permission
	authPayload := ctx.MustGet(authPayloadKey).(*auth.Payload)
	permit, err := server.storage.GetPermission(ctx, authPayload.Subject)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if reqUri.ID != authPayload.Subject || permit.Role != "admin" {
		ctx.JSON(http.StatusForbidden, errorResponse(ErrAccessDenied))
		return
	}

	arg := db.UpdateUserParams{
		ID: reqUri.ID,
		Username: reqJSON.Username,
		Email: reqJSON.Email,
	}

	user, err := server.storage.UpdateUser(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, user)
}