package api

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/hasnaroihan/grocery-planner/auth"
	db "github.com/hasnaroihan/grocery-planner/db/sqlc"
)

var SYM_KEY string
var ACCESS_TOKEN_DURATION time.Duration

var (
	ErrAccessDenied = errors.New("authenticated user does not have access permission")
)

type Server struct {
	storage       db.Storage
	tokenMaker    auth.TokenMaker
	tokenDuration time.Duration
	router        *gin.Engine
}

// Server constructor
func NewServer(storage db.Storage) (*Server, error) {
	err := configToken()
	if err != nil {
		return nil, err
	}

	tokenMaker, err := auth.NewPASETOToken(SYM_KEY)
	if err != nil {
		return nil, fmt.Errorf("unable to create PASETO maker: %s", err)
	}
	server := &Server{
		storage:       storage,
		tokenMaker:    tokenMaker,
		tokenDuration: ACCESS_TOKEN_DURATION,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("username", usernameValidator)
	}

	// add routes to the router
	server.setupRouter()

	return server, nil
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func (server *Server) setupRouter() {
	router := gin.Default()
	authRouter := router.Group("/").Use(authMiddleware(server.tokenMaker))
	adminRouter := router.Group("/").Use(authMiddleware(server.tokenMaker), adminMiddleware(server.storage))

	// USER
	router.POST("/register", server.registerUser)
	router.POST("/login", server.loginUser)
	adminRouter.DELETE("/user/delete/:id", server.deleteUser)
	adminRouter.GET("/user/all", server.listUsers)
	authRouter.GET("/user/:id", server.getUser)
	authRouter.PATCH("/user/update/:id", server.updateUser)
	// TODO update verified and update password

	// INGREDIENTS
	adminRouter.POST("/ingredients/add", server.createIngredient)
	adminRouter.DELETE("/ingredients/delete/:id", server.deleteIngredient)
	adminRouter.PATCH("/ingredients/update/:id", server.updateIngredient)
	// authRouter.GET("/ingredients/:id", server.getIngredient)
	// authRouter.GET("/ingredients/all", server.listIngredients)
	// authRouter.GET("/ingredients", server.searchIngredients)
	router.GET("/ingredients/:id", server.getIngredient)
	router.GET("/ingredients/all", server.listIngredients)
	router.GET("/ingredients", server.searchIngredients)

	// UNITS
	adminRouter.POST("/unit/add", server.createUnit)
	adminRouter.DELETE("/unit/delete/:id", server.deleteUnit)
	adminRouter.PATCH("/unit/update/:id", server.updateUnit)
	authRouter.GET("/unit/:id", server.getUnit)
	authRouter.GET("/unit/all", server.listUnits)

	// RECIPES
	authRouter.POST("/recipe/add", server.newRecipe)
	authRouter.DELETE("/recipe/delete/:id", server.deleteRecipe)
	authRouter.DELETE("/recipe/delete", server.deleteRecipeIngredient)
	authRouter.PATCH("/recipe/update/:id", server.updateRecipe)
	authRouter.GET("/recipe/my", server.listRecipesUser)
	router.GET("/recipe/:id", server.getRecipe)
	router.GET("/recipe/all", server.listRecipes)
	router.GET("/recipe", server.searchRecipe)

	// SCHEDULES
	router.POST("/groceries", server.generateGroceries)
	adminRouter.GET("/schedule/all", server.listSchedules)
	authRouter.GET("/schedule/list", server.listSchedulesUser)
	authRouter.DELETE("/schedule/delete/:id", server.deleteSchedule)
	authRouter.DELETE("/schedule/delete", server.deleteScheduleRecipe)

	server.router = router
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func configToken() error {
	minDuration, err := strconv.Atoi(os.Getenv("ACCESS_TOKEN_DURATION"))
	if err != nil {
		return fmt.Errorf("error loading environment variables. err: %s", err)
	}

	SYM_KEY = os.Getenv("SYM_KEY")
	ACCESS_TOKEN_DURATION = time.Duration(time.Duration(minDuration) * time.Minute)

	return nil
}
