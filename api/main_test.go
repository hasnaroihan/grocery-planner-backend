package api

import (
	"log"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	db "github.com/hasnaroihan/grocery-planner/db/sqlc"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

const envPath = "./../.env"
func newTestServer(t *testing.T, storage db.Storage) *Server {
	err := godotenv.Load(envPath)
	if err != nil {
		log.Fatalf("Error loading environment variables. Err: %s", err)
	}
	
	server, err := NewServer(storage)
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}