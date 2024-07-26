package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/hasnaroihan/grocery-planner/api"
	db "github.com/hasnaroihan/grocery-planner/db/sqlc"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var POSTGRES_USER string
var POSTGRES_PASSWORD string
var POSTGRES_HOST string
var SERVER_ADDRESS string
var SYM_KEY string
var dbDriver string
var dbSource string

const (
	envPath = ".env"
)

func main() {
	initConfig()

	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("Cannot connect to the database", err)
	}

	storage := db.NewStorage(conn)
	server, err := api.NewServer(storage)
	if err != nil {
		log.Fatal("Cannot create server", err)
	}

	SERVER_ADDRESS = os.Getenv(SERVER_ADDRESS)
	err = server.Start(SERVER_ADDRESS)
	if err != nil {
		log.Fatal("Cannot start server", err)
	}
}

func initConfig() {
	err := godotenv.Load(envPath)
	if err != nil {
		log.Fatalf("Error loading environment variables. Err: %s", err)
	}
	POSTGRES_USER = os.Getenv("POSTGRES_USER")
	POSTGRES_PASSWORD = os.Getenv("POSTGRES_PASSWORD")
	POSTGRES_HOST = os.Getenv("HOST")
	SYM_KEY = os.Getenv("SYM_KEY")

	dbDriver = "postgres"
	dbSource = fmt.Sprintf("postgresql://%s:%s@%v:5432/grocery-planner?sslmode=disable",
		POSTGRES_USER,
		POSTGRES_PASSWORD,
		POSTGRES_HOST)
}
