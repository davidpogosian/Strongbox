package files

import (
	"net/http"
	"strings"
	"strongbox/platform/storage"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type File struct {
	Name         string
	S3Key        string
	Size         int64
	LastModified string
}

type Folder struct {
	Name string
	S3Key string
}

func Handler(s3Client *s3.Client) gin.HandlerFunc {
	return func (ctx *gin.Context) {
		session := sessions.Default(ctx)
		profile := session.Get("profile").(map[string]interface{})
		userId := profile["sub"].(string)

		// Extract folder path from URL
        path := ctx.Param("path")
        if path == "/" {
        	path = ""
        }

        // List objects in S3 with the user ID as the prefix
        trunk := userId + path + "/"
		objects, err := storage.ListObjects(s3Client, trunk)
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		var files []File
		foldersSet := make(map[Folder]bool)
        var folders []Folder
        for _, object := range objects {
            trimmedKey := strings.TrimPrefix(*object.Key, trunk)
            parts := strings.Split(trimmedKey, "/")
            if len(parts) == 1 {
		        name := parts[0]
		        files = append(files, File{
		            Name:         name,
		            S3Key:        *object.Key,
		            Size:         *object.Size,
		            LastModified: object.LastModified.Format(time.RFC3339),
		        })
            } else {
            	folder := Folder{
             		Name: parts[0],
               		S3Key: trunk + parts[0],
             	}
            	foldersSet[folder] = true
            }
        }

        for folder := range foldersSet {
        	folders = append(folders, folder)
        }

        data := gin.H{
            "profile":     	profile,
            "files": 		files,
            "folders":      folders,
            "currentPath":  path,
        }

        ctx.HTML(http.StatusOK, "files.html", data)
	}
}
