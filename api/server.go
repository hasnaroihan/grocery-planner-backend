package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/hasnaroihan/grocery-planner/db/sqlc"
)

type Server struct {
	storage db.Storage
	router *gin.Engine
}

// Server constructor
func NewServer(storage db.Storage) *Server {
	server := &Server{
		storage: storage,
	}
	router := gin.Default()

	// add routes to the router
	// USER (postponed until i understand how to implement hash, salt, and auth)
	router.POST("/register", server.registerUser)
	//router.POST("/login", server.loginUser)

	// INGREDIENTS
	router.POST("/ingredients/add", server.createIngredient)
	router.DELETE("/ingredients/delete/:id", server.deleteIngredient)
	router.GET("/ingredients/:id", server.getIngredient)
	router.GET("/ingredients/all", server.listIngredients)
	router.GET("/ingredients", server.searchIngredients)
	router.PATCH("/ingredients/update/:id", server.updateIngredient)

	// UNITS
	router.POST("/unit/add", server.createUnit)
	router.DELETE("/unit/delete/:id", server.deleteUnit)
	router.GET("/unit/:id", server.getUnit)
	router.GET("/unit/all", server.listUnits)
	router.PATCH("/unit/update/:id", server.updateUnit)

	// RECIPES
	router.POST("/recipe/add", server.newRecipe)
	router.DELETE("/recipe/delete/:id", server.deleteRecipe)
	router.DELETE("/recipe/delete", server.deleteRecipeIngredient)
	router.GET("/recipe/:id", server.getRecipe)
	router.GET("/recipe/all", server.listRecipes)
	router.GET("/recipe", server.searchRecipe)
	router.PATCH("/recipe/update/:id", server.updateRecipe)

	// SCHEDULES
	router.POST("/groceries", server.generateGroceries)
	router.GET("/schedule", server.listSchedules)
	router.DELETE("/schedule/delete/:id", server.deleteSchedule)
	router.DELETE("/schedule/delete", server.deleteScheduleRecipe)

	server.router = router

	return server
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error" : err.Error()}
}