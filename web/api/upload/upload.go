package upload

import (
	"io"
	"net/http"
	"strongbox/platform/storage"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Handler handles the file upload
func Handler(s3Client *s3.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		profile := session.Get("profile").(map[string]interface{})
		userId := profile["sub"].(string)

		form, err := ctx.MultipartForm()
		if err != nil {
			ctx.String(http.StatusBadRequest, "Unable to parse form data: %s", err.Error())
			return
		}

		files := form.File["files[]"] // Retrieve all files from the "files[]" field
		filePaths := form.Value["filePaths[]"]

		for i, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				ctx.String(http.StatusInternalServerError, "Unable to open file: %s", err.Error())
				return
			}
			defer file.Close()

			// Read the file into a byte slice
			fileBytes, err := io.ReadAll(file)
			if err != nil {
				ctx.String(http.StatusInternalServerError, "Unable to read file: %s", err.Error())
				return
			}

			// Construct the S3 key based on the user ID and the file's relative path
			key := userId + "/" + filePaths[i]

			// Atomic: Put file in S3, check if already in S3, ask to overwrite
			overwrite := ctx.Query("overwrite") == "true"
			exists, err := storage.ObjectExists(s3Client, key)
			if err != nil {
				ctx.String(http.StatusInternalServerError, "Unable to check if file already exists: %s", err.Error())
				return
			}
			if exists && !overwrite {
				ctx.String(http.StatusConflict, "Asset with the name '%s' already exists", fileHeader.Filename)
				return
			}

			// Upload the file to S3
			err = storage.UploadFile(s3Client, key, fileBytes)
			if err != nil {
				ctx.String(http.StatusInternalServerError, "Unable to upload file: %s", err.Error())
				return
			}
		}

		// Respond to the client after all files are successfully uploaded
		ctx.JSON(http.StatusOK, gin.H{"message": "Files uploaded successfully"})
	}
}
