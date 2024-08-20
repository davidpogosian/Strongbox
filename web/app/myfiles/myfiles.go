package myfiles

import (
	"database/sql"
	"net/http"
	"strings"
	"strongbox/platform/database"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// TODO need S3 key in this struct
type VerboseAsset struct {
	Name string
	/* download link? */
	/* size? */
}

func Handler(db *sql.DB) gin.HandlerFunc {
	return func (ctx *gin.Context) {
		session := sessions.Default(ctx)
		profile := session.Get("profile").(map[string]interface{})

		// Get files, fetch them from S3, display metadata
		assets, err := database.FindAllAssetsByUserId(db, profile["sub"].(string))
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		var verboseAssets []VerboseAsset
		for _, asset := range assets {
			parts := strings.Split(asset.S3Key, "/")
			name := parts[len(parts) - 1]
			verboseAssets = append(verboseAssets, VerboseAsset{Name: name,})
		}

		data := gin.H{
			"profile": profile,
			"verboseAssets": verboseAssets,
		}

		ctx.HTML(http.StatusOK, "myfiles.html", data)
	}
}
