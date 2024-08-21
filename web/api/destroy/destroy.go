package destroy

import (
	"net/http"
	"strings"
	"strongbox/platform/storage"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func Handler(s3Client *s3.Client) gin.HandlerFunc {
	return func (ctx *gin.Context) {
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

		// List all objects under the key prefix
		objects, err := storage.ListObjects(s3Client, key)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Unable to list objects: %s", err.Error())
			return
		}

		// Delete all objects
		for _, object := range objects {
			err := storage.DestroyObject(s3Client, *object.Key)
			if err != nil {
				ctx.String(http.StatusInternalServerError, "Unable to delete object: %s", err.Error())
				return
			}
		}

		// Respond with success
		ctx.JSON(http.StatusOK, gin.H{"message": "Files and folders deleted successfully"})
	}
}
