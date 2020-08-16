package vkore

import (
	"encoding/json"
	"fmt"
	vkapi "github.com/himidori/golang-vk-api"
	"github.com/vkore/vkore/internal/store"
	"github.com/vkore/vkore/pkg/vkapi/models"
	"log"
	"net/url"
	"os"
	"sync"
	"time"
)

var groupsToParse = []string{"podolsk_naodinraz", "knowledge50pd", "znakomstva_v_podolske", "virtual.dating", "podolsk_love", "znakomstvoodolsk", "podolsk_znakomstva_v", "publicpoznakomlys2016"}

//var groupsToParse = []string{"knowledge50pd", "podolsk_love", "virtual.dating", "podolsk_naodinraz", "znakomstva_v_podolske", "public192002298"}

var wg sync.WaitGroup

func Do() {
	client, err := vkapi.NewVKClient(vkapi.DeviceIPhone, os.Getenv("VK_USER"), os.Getenv("VK_PASSWORD"), true)

	if err != nil {
		log.Fatal(err)
	}

	for _, group := range groupsToParse {
		groupInfo, err := ResolveScreenName(client, group)
		if err != nil {
			log.Println(err)
		}

		g := &models.Group{
			ID:         groupInfo.ObjectID,
			ScreenName: group,
			Type:       groupInfo.Type,
		}

		groupLastUpdate, err := store.GetGroupLastUpdate(groupInfo.ObjectID)
		if err == nil {
			yesterday := time.Now().AddDate(0, 0, -1)
			if groupLastUpdate == nil {
				log.Printf(`Last update for group "%v" is nil and no errors`, group)
			} else if groupLastUpdate.After(yesterday) {
				log.Printf(`no need to update for group "%v": %v`, group, groupLastUpdate.Format(time.RFC822))
				continue
			}
		}

		groupMembers, err := GetGroupMembers(client, g)
		if err != nil {
			log.Println("can't get members:", groupMembers)
		}
	}
	wg.Wait()

	filters := []*store.Filter{
		{Query: models.User{Sex: 1}},
		{Query: "deactivated IS NULL"},
		{Query: "last_seen > ?", Args: []interface{}{time.Now().AddDate(0, 0, -4)}},
	}

	target := store.GetUsers(filters...)

	fmt.Println("GOT USERS", len(target))
}

func GetGroupMembers(c *vkapi.VKClient, group *models.Group) ([]*models.User, error) {
	var users []*models.User
	if err := store.GormDB().Where(models.Group{ID: group.ID}).Attrs(group).FirstOrCreate(group).Error; err != nil {
		log.Printf("error get or create gorup in database: %v", err)
		return nil, err
	}

	totalCount := 25000
	offset := 0
	for {
		values := make(url.Values)
		values.Set("code", fmt.Sprintf(`
var members = [];
var offset = 0;

var count = 0;
var i = 0;
while (i < 25 && (offset + %v) < %v) {
  var m = API.groups.getMembers({
    "group_id": %v,
    "v": "5.27",
    "sort": "id_asc",
    "count": "1000",
    "offset": (%v + offset),
    "fields": "sex,deactivated,last_seen"
  });
  members.push(m.items);
  count = m.count;
  offset = offset + 1000;
  i = i + 1;
};
return { "users": members, "count": count };
`, offset, totalCount, group.ID, offset))

		r, err := c.MakeRequest("execute", values)
		if err != nil {
			log.Println("request error", err)
			return nil, err
		}
		//resp := new(models.GroupMembers)
		//var resp interface{}
		var resp struct {
			Users [][]*models.User `json:"users"`
			Count int              `json:"count"`
		}
		err = json.Unmarshal(r.Response, &resp)
		if err != nil {
			log.Println("error unmarshal")
			return nil, err
		}

		for _, us := range resp.Users {
			users = append(users, us...)
		}

		_ = store.CreateUsers(users)
		err = store.AddGroupMembers(group.ID, users)
		if err != nil {
			log.Println("error creating users in database:", err)
		}
		fmt.Println("users added")
		if len(users) >= resp.Count {
			break
		}
		offset += 25000
		totalCount += 25000
	}

	return users, nil
}

func ResolveScreenName(c *vkapi.VKClient, screenName string) (*models.Resolver, error) {
	values := make(url.Values)
	values.Set("screen_name", screenName)

	r, err := c.MakeRequest("utils.resolveScreenName", values)
	if err != nil {
		return nil, err
	}

	resp := new(models.Resolver)
	err = json.Unmarshal(r.Response, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func MigrateSchema() {
	store.GormDB().AutoMigrate(&models.User{}, &models.Group{})
}