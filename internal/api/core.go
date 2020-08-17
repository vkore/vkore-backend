package api

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/vkore/vkore/internal/store"
	"github.com/vkore/vkore/pkg/vkapi/models"
	"net/http"
	"time"
)

func ListenAndServe() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Use(cors.Default())
	r.GET("/pages", GetUsers)
	r.GET("/api/all_groups", GetAllGroups)
	r.Run()
}

//func GetPages(c *gin.Context) {
//
//	//var s map[string]interface{}
//	//err := c.Bind(&s)
//	//if err != nil {
//	//	c.JSON(http.StatusBadRequest, gin.H{
//	//		"error": fmt.Sprintf(`%v`, err),
//	//	})
//	//}
//
//	r := vkore.GetPages()
//	if len(r) > 20 {
//		c.JSON(http.StatusOK, vkore.GetPages()[:20])
//		return
//	}
//
//	c.JSON(http.StatusOK, vkore.GetPages())
//}

func GetUsers(c *gin.Context) {
	filters := []*store.Filter{
		{Query: models.User{Sex: 1}},
		{Query: "deactivated IS NULL"},
		{Query: "last_seen > ?", Args: []interface{}{time.Now().AddDate(0, 0, -4)}},
	}

	target := store.GetUsers(filters...)
	if len(target) > 20 {
		c.JSON(http.StatusOK, target[:50])
		return
	}
}

func ParseGroups(c *gin.Context) {
	var groups []string

	err := c.Bind(&groups)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": fmt.Sprintf("%v", err),
		})
	}

	// ["http://vk.com/groupname", "http://vk.com/groupname2"]...
}

func GetAllGroups(c *gin.Context) {

	groups, err := store.GetAllGroups()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("%v", err),
		})
		return
	}

	c.JSON(http.StatusOK, groups)
}
