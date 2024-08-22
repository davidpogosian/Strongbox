package downloadFolder

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"strongbox/platform/storage"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// zipObjects creates a zip file in memory containing the provided objects
func zipObjects(s3Client *s3.Client, objects []types.Object, prefixToRemove string) (*bytes.Buffer, error) {
	// Create a buffer to store the zip file
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Iterate over the objects to add them to the zip
	for _, obj := range objects {
		// Open the object from S3 (you'll need to implement GetObject)
		data, err := storage.GetObject(s3Client, *obj.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve object: %w", err)
		}

		// Create a file in the zip archive
		relativePath := strings.TrimPrefix(*obj.Key, prefixToRemove)
		fileWriter, err := zipWriter.Create(relativePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file in zip: %w", err)
		}

		// Copy the object's data into the zip file
		_, err = io.Copy(fileWriter, data)
		if err != nil {
			return nil, fmt.Errorf("failed to write object data to zip: %w", err)
		}
	}

	// Close the zip writer to finalize the archive
	err := zipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	return buf, nil
}

func Handler(s3Client *s3.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		profile := session.Get("profile").(map[string]interface{})
		userId := profile["sub"].(string)

		key := ctx.Query("key")
		if key == "" {
			ctx.String(http.StatusBadRequest, "No key in url")
			return
		}

		// Authorize user
		idFromKey := strings.Split(key, "/")[0]
		if idFromKey == "" {
			ctx.String(http.StatusBadRequest, "Malformed key")
			return
		}

		if idFromKey != userId {
			ctx.String(http.StatusForbidden, "Not your asset")
			return
		}

		// List all objects with this key prefix
		objects, err := storage.ListObjects(s3Client, key)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Failed to list objects: %s", err)
			return
		}

		// Zip them into a single file
		parts := strings.Split(key, "/")
		folderName := parts[len(parts) - 1]
		partToRemove := strings.TrimSuffix(key, folderName)
		zipBuffer, err := zipObjects(s3Client, objects, partToRemove)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Failed to create zip: %s", err)
			return
		}

		// Convert the buffer to []byte
		zipData, err := io.ReadAll(zipBuffer)
		if err != nil {
		    ctx.String(http.StatusInternalServerError, "Failed to read zip buffer: %s", err)
		    return
		}

		// Upload zipped file
		zippedKey := key + ".zip"
		err = storage.UploadFile(s3Client, zippedKey, zipData)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Failed to upload zip file: %s", err)
			return
		}

		// Create a pre-signed URL for the zipped file
		link, err := storage.GeneratePresignedURL(s3Client, zippedKey, time.Hour*24)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Unable to generate download link: %s", err)
			return
		}

		// Send the link to the user
		ctx.JSON(http.StatusOK, gin.H{"download_url": link})
	}
}
