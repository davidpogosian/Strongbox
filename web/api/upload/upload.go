package upload

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Handler handles the file upload
func Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		profile := session.Get("profile")
		fmt.Println(profile)

		// Get the file from the request
		file, fileHeader, err := ctx.Request.FormFile("file")
		if err != nil {
			ctx.String(http.StatusBadRequest, "Unable to get file: %s", err.Error())
			return
		}
		defer file.Close()

		// Print the file details
		fmt.Println("Received file:", fileHeader.Filename)

		// Respond to the client
		ctx.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
	}
}
