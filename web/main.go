package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Wsine/feishu2md/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	if mode := os.Getenv("GIN_MODE"); mode != "release" {
		utils.LoadEnv()
	}

	router := gin.New()
	router.LoadHTMLGlob(utils.RootDir() + "/web/templ/*.templ.html")
	router.Static("/static", utils.RootDir()+"/web/static")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.templ.html", nil)
	})
	router.GET("/download", downloadHandler)

	if err := router.Run(); err != nil {
		log.Panicf("error: %s", err)
	}
}
