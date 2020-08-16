package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/vkore/vkore/internal/vkore"
	"net/http"
)

func ListenAndServe() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Use(cors.Default())
	r.GET("/pages", GetPages)
	r.Run()
}

func GetPages(c *gin.Context) {

	//var s map[string]interface{}
	//err := c.Bind(&s)
	//if err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"error": fmt.Sprintf(`%v`, err),
	//	})
	//}

	r := vkore.GetPages()
	if len(r) > 20 {
		c.JSON(http.StatusOK, vkore.GetPages()[:20])
		return
	}

	c.JSON(http.StatusOK, vkore.GetPages())
}
