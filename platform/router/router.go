package router

import (
	"database/sql"
	"encoding/gob"
	"encoding/hex"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"strongbox/platform/authenticator"
	"strongbox/platform/middleware"
	"strongbox/web/api/upload"
	"strongbox/web/api/download"
	"strongbox/web/api/downloadFolder"
	"strongbox/web/api/destroy"
	"strongbox/web/app/callback"
	"strongbox/web/app/home"
	"strongbox/web/app/login"
	"strongbox/web/app/logout"
	"strongbox/web/app/files"
)

// New registers the routes and returns the router.
func New(db *sql.DB, auth *authenticator.Authenticator, s3Client *s3.Client) *gin.Engine {
	router := gin.Default()

	// To store custom types in our cookies,
	// we must first register them using gob.Register
	gob.Register(map[string]interface{}{})

	key, err := hex.DecodeString(os.Getenv("COOKIE_KEY"))
	if err != nil {
		log.Fatal("Can't decode COOKIE_KEY:", err)
	}
	store := cookie.NewStore(key)
	router.Use(sessions.Sessions("auth-session", store))

	router.Static("/public", "web/static")
	router.LoadHTMLGlob("web/template/*")

	router.GET("/", home.Handler)
	router.GET("/login", login.Handler(auth))
	router.GET("/callback", callback.Handler(auth))
	router.GET("/files/*path", middleware.IsAuthenticated, files.Handler(s3Client))
	router.GET("/logout", logout.Handler)

	api := router.Group("/api")
	{
		api.POST("/upload", middleware.MustBeAuthenticated, upload.Handler(s3Client))
		api.GET("/download", middleware.MustBeAuthenticated, download.Handler(s3Client))
		api.DELETE("/destroy", middleware.MustBeAuthenticated, destroy.Handler(s3Client))
		api.GET("/downloadFolder", middleware.MustBeAuthenticated, downloadFolder.Handler(s3Client))
	}

	return router
}
