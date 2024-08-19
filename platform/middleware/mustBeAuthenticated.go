package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func MustBeAuthenticated(ctx *gin.Context) {
	if sessions.Default(ctx).Get("profile") == nil {
		ctx.String(http.StatusForbidden, "Not authenticated")
		ctx.Abort()
	} else {
		ctx.Next()
	}
}
