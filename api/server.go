package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/hasnaroihan/grocery-planner/db/sqlc"
)

type Server struct {
	storage *db.Storage
	router *gin.Engine
}

// Server constructor
func NewServer(storage *db.Storage) *Server {
	server := &Server{
		storage: storage,
	}
	router := gin.Default()

	// add routes to the router
	router.POST("/register", server.registerUser)

	server.router = router

	return server
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error" : err.Error()}
}