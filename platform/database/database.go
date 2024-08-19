package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Asset struct {
	UserId string
	S3Key string
	LastAccessed time.Time
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
		s3Key TEXT NOT NULL,
		lastAccessed TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.Exec(statement)
	if err != nil {
		log.Fatal(err)
	}
}

func AddAsset(db *sql.DB, asset *Asset) {
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

func FindAllAssetsByUserId(db *sql.DB, userId string) ([]Asset, error) {
	query := "SELECT userId, s3Key, lastAccessed FROM assets WHERE userId = ?"
    rows, err := db.Query(query, userId)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var assets []Asset
    for rows.Next() {
        var asset Asset
        err := rows.Scan(&asset.UserId, &asset.S3Key, &asset.LastAccessed)
        if err != nil {
            return nil, err
        }
        assets = append(assets, asset)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return assets, nil
}

func DeleteDatabase(name string) {
	err := os.Remove(fmt.Sprintf("./%s", name))
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}
}
