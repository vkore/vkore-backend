package api

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/vkore/vkore/internal/store"
	"github.com/vkore/vkore/internal/vkore"
	"github.com/vkore/vkore/pkg/vkapi/models"
	"net/http"
	"regexp"
	"time"
)

var vkURL = regexp.MustCompile(`^((https?:\/\/)?(m\.)?((vkontakte|vk)\.)(com|ru)\/)?(?P<group>[0-9a-zA-Z_-]+)$`)

func ListenAndServe() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Use(cors.Default())
	r.POST("/pages", GetUsers)
	r.GET("/api/all_groups", GetAllGroups)
	r.POST("/api/load_groups", LoadGroups)
	r.GET("/api/get_cities", getCities)
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

func invalidVariable(c *gin.Context, name string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": fmt.Sprintf(`invalid query param "%v"`, name),
	})
}

type Pagination struct {
	PerPage int `json:"perPage"`
	Offset  int `json:"offset"`
}

type Filters struct {
	Cities   []int `json:"cities"`
	LastSeen *int  `json:"lastSeen"`
}

func rErr(c *gin.Context, err error, status int) {
	c.JSON(status, gin.H{
		"error": err.Error(),
	})
}

// GetUsers - получение пользователей из БД. Главная страница
func GetUsers(c *gin.Context) {
	type request struct {
		Pagination Pagination `json:"pagination"`
		Filters    Filters    `json:"filters"`
	}
	req := new(request)

	err := c.BindJSON(req)
	if err != nil {
		fmt.Println("err", err)
		rErr(c, err, http.StatusBadRequest)
		return
	}

	filters := []*store.Filter{
		{Query: models.User{Sex: 1}},
		{Query: "deactivated IS NULL"},
	}

	if len(req.Filters.Cities) != 0 {
		var vvv interface{}
		vvv = req.Filters.Cities

		filters = append(filters, &store.Filter{
			Query: "city_id in (?)",
			Args:  []interface{}{vvv},
		})
	}

	if req.Filters.LastSeen != nil {
		fmt.Println("LAST SEEN", *req.Filters.LastSeen)
		filters = append(filters, &store.Filter{
			Query: "last_seen > ?",
			Args:  []interface{}{time.Now().AddDate(0, 0, -*req.Filters.LastSeen)},
		})
	}

	target, totalCount := store.GetUsers(req.Pagination.PerPage, req.Pagination.Offset, filters...)
	c.JSON(http.StatusOK, gin.H{
		"items":      target,
		"totalCount": totalCount,
	})
}

func LoadGroups(c *gin.Context) {
	var groups []string

	err := c.Bind(&groups)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": fmt.Sprintf("%v", err),
		})
	}

	var groupNames []string
	for _, group := range groups {
		match := vkURL.FindStringSubmatch(group)
		for j, name := range vkURL.SubexpNames() {
			if name == "group" && len(match) > j && match[j] != "" {
				groupNames = append(groupNames, match[j])
			}
		}
	}

	vkore.GetPages(groupNames)

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

func getCities(c *gin.Context) {

	cities, err := store.GetAllCities()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("%v", err),
		})
	}

	c.JSON(http.StatusOK, cities)
}
