package myfiles

import (
	"database/sql"
	"fmt"
	"net/http"
	"strongbox/platform/database"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

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
		fmt.Println(assets)

		ctx.HTML(http.StatusOK, "myfiles.html", profile)
	}
}
