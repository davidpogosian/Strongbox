package upload

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strongbox/platform/database"
	"strongbox/platform/storage"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Handler handles the file upload
func Handler(db *sql.DB, s3Client *s3.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		profile := session.Get("profile").(map[string]interface{})
		userId := profile["sub"].(string)

		// Get the file from the request
		file, fileHeader, err := ctx.Request.FormFile("file")
		if err != nil {
			ctx.String(http.StatusBadRequest, "Unable to get file: %s", err.Error())
			return
		}
		defer file.Close()

		// Print the file details
		fmt.Println("Received file:", fileHeader.Filename)

		// Read the file into a byte slice
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Unable to read file: %s", err.Error())
			return
		}

		// Atomic: Put file in S3, check if already in S3, ask to overwrite, put file in db
		// TODO use storage.ObjectExists
		overwrite := ctx.Query("overwrite") == "true"
		key := userId + "/" + fileHeader.Filename
		exists, err := storage.ObjectExists(s3Client, key)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Unable to check if file already exists: %s", err.Error())
			return
		}
		if exists && !overwrite {
			ctx.String(http.StatusConflict, "Asset with this name already exists")
			return
		}

		err = storage.UploadFile(s3Client, key, fileBytes)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Unable to upload file: %s", err.Error())
			return
		}

		err = database.AddAsset(db, &database.Asset{
			UserId: userId,
			S3Key: key,
		})
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Unable to put file in database: %s", err.Error())
			// TODO Delete asset from S3
			err = storage.DeleteFile(s3Client, key)
			if err != nil {
				// log something somewhere so when S3 comes back online we can delete
			}
			return
		}

		// Respond to the client
		ctx.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
	}
}
