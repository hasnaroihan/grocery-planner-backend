package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var testQueries *Queries
var POSTGRES_USER string
var POSTGRES_PASSWORD string
var HOST string
var dbDriver string
var dbSource string
const envPath = "./../../.env"

func TestMain(m *testing.M) {
	InitTestConfig()
	
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("Cannot connect to the database", err)
	}

	testQueries = New(conn)
	
	os.Exit(m.Run())
}

func InitTestConfig() {
	err := godotenv.Load(envPath)
	if err != nil {
		log.Fatalf("Error loading environment variables. Err: %s", err)
	}
	POSTGRES_USER = os.Getenv("POSTGRES_USER")
	POSTGRES_PASSWORD = os.Getenv("POSTGRES_PASSWORD")
	HOST = os.Getenv("HOST")

	dbDriver = "postgres"
	dbSource = fmt.Sprintf("postgresql://%s:%s@%v:5432/grocery-planner?sslmode=disable", 
				POSTGRES_USER, 
				POSTGRES_PASSWORD, 
				HOST)
}