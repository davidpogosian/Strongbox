package download

import (
	"net/http"
	"strings"
	"strongbox/platform/storage"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

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

		// Create link
		link, err := storage.GeneratePresignedURL(s3Client, key, time.Hour * 24)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Unable to generate download link: %s", err)
			return
		}

		// send link to user
		ctx.JSON(http.StatusOK, gin.H{"download_url": link})
	}
}
