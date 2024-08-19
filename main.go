package main

import (
	"log"
	"net/http"

	"strongbox/platform/authenticator"
	"strongbox/platform/database"
	"strongbox/platform/router"
	"strongbox/platform/storage"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load the env vars: %v", err)
	}

	database.DeleteDatabase("strongbox.db")
	db := database.InitializeDatabaseConnection("strongbox.db")
	defer database.TeardownDatabaseConnection(db)
	database.CreateAssetTable(db)

	auth, err := authenticator.New()
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	s3Client := storage.InitializeStorage()

	rtr := router.New(db, auth, s3Client)

	log.Print("Server listening on http://localhost:3000/")
	if err := http.ListenAndServe("0.0.0.0:3000", rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}
