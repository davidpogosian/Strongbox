package database

import (
	"os"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Asset struct {
	UserId string
	S3Key string
}

func InitializeDatabaseConnection(name string) *sql.DB {
	db, err := sql.Open("sqlite3", fmt.Sprintf("./%s", name))
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func TeardownDatabaseConnection(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Fatal(err)
	}
}

func CreateAssetTable(db *sql.DB) {
	statement := `
	CREATE TABLE IF NOT EXISTS assets (
		userId TEXT NOT NULL,
		s3Key TEXT NOT NULL
	);
	`
	_, err := db.Exec(statement)
	if err != nil {
		log.Fatal(err)
	}
}

func AddWord(db *sql.DB, asset *Asset) {
	statement, err := db.Prepare("INSERT INTO assets (userId, s3Key) VALUES (?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer statement.Close()
	_, err = statement.Exec(asset.UserId, asset.S3Key)
	if err != nil {
		log.Fatal(err)
	}
}

func DeleteDatabase(name string) {
	err := os.Remove(fmt.Sprintf("./%s", name))
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}
}
