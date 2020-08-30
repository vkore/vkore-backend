package vkore

import (
	"bytes"
	"encoding/json"
	"fmt"
	vkapi "github.com/himidori/golang-vk-api"
	"github.com/vkore/vkore/internal/store"
	"github.com/vkore/vkore/pkg/vkapi/models"
	"log"
	"net/url"
	"os"
	"runtime/debug"
	"sync"
	"text/template"
	"time"
)

var client *vkapi.VKClient

func Init() {
	var err error
	client, err = vkapi.NewVKClient(vkapi.DeviceIPhone, os.Getenv("VK_USER"), os.Getenv("VK_PASSWORD"), true)
	if err != nil {
		log.Fatal(err)
	}
	client.Client.Timeout = 240 * time.Second
}

type groupMembersParams struct{ Offset, TotalCount, GroupID, UsersCount, PerPage int }

func (gmp groupMembersParams) NextPage() *groupMembersParams {
	gmp.Offset += gmp.PerPage
	gmp.TotalCount += gmp.PerPage
	return &gmp
}

func GetPages(groupsToParse []string) {
	for _, group := range groupsToParse {
		var g *models.Group
		rrr, err := store.GetGroupByScreenName(group)
		if rrr == nil || err != nil {
			groupInfo, err := ResolveScreenName(group)
			if err != nil {
				log.Printf("can't get info about %v group: %v", group, err)
				continue
			}
			g = &models.Group{
				ID:         groupInfo.ObjectID,
				ScreenName: group,
				Type:       groupInfo.Type,
			}
		} else {
			g = rrr
		}

		groupMembers, err := GetGroupMembers(g)
		log.Println("GROUP MEMBERS:", len(groupMembers))
		if err != nil {
			log.Println("can't get members:", groupMembers)
		}
		groupMembers = nil
	}
	fmt.Println("FREE MEMORY")
	debug.FreeOSMemory()
}

func groupMembersQuery(gmp *groupMembersParams) string {
	t := template.New("Get group users")

	params := struct {
		*groupMembersParams
		Loops int
	}{gmp, gmp.PerPage / 1000}

	_, _ = t.Parse(`
var members = [];
var offset = 0;

var count = 0;
var i = 0;
while (i < {{.Loops}} && (offset + {{.Offset}}) < {{.TotalCount}}) {
  var m = API.groups.getMembers({
    "group_id": {{.GroupID}},
    "v": "5.27",
    "sort": "id_asc",
    "count": "{{.UsersCount}}",
    "offset": ({{.Offset}} + offset),
    "fields": "sex,deactivated,last_seen,photo,photo_200,city,status"
  });
  members.push(m.items);
  count = m.count;
  offset = offset + {{.UsersCount}};
  i = i + 1;
};
return { "users": members, "count": count };
`)
	var w bytes.Buffer
	_ = t.Execute(&w, params)

	return w.String()
}

func getGroupMembers(queryParams *groupMembersParams, result *userGroupsResult) error {
	values := make(url.Values)
	values.Set("code", groupMembersQuery(queryParams))

	r, err := client.MakeRequest("execute", values)
	if err != nil {
		log.Println("request error", err)
		return err
	}
	fmt.Println("RESPONSE ERROR", r.ResponseError)

	err = json.Unmarshal(r.Response, result)
	if err != nil {
		log.Println("error unmarshal", err)
		return err
	}
	uss := 0
	for _, user := range result.Users {
		uss += len(user)
	}
	fmt.Println("GOTTTT", uss, "USERS")
	return nil
}

type userGroupsResult struct {
	Users [][]*models.User `json:"users"`
	Count int              `json:"count"`
}

func GetGroupMembers(group *models.Group) ([]*models.User, error) {
	count, _, err := client.GroupGetMembers(group.ID, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("error getting group members count: %v", err)
	}
	users := make(chan *models.User, count)

	if err := store.GormDB().Where(models.Group{ID: group.ID}).Attrs(group).FirstOrCreate(group).Error; err != nil {
		log.Printf("error get or create gorup in database: %v", err)
		return nil, err
	}

	queryParams := groupMembersParams{
		Offset:     0,
		TotalCount: 10000,
		GroupID:    group.ID,
		UsersCount: 1000,
		PerPage:    10000,
	}

	ticker := time.NewTicker(334 * time.Millisecond)

	var wg sync.WaitGroup
	for {
		<-ticker.C
		wg.Add(1)
		go func(qp groupMembersParams) {
			defer wg.Done()
			fmt.Println("START QUERY")
			var resp userGroupsResult
			err := getGroupMembers(&qp, &resp)
			if err != nil {
				log.Println("error getting group memebers:", err)
			}
			for _, us := range resp.Users {
				for _, u := range us {
					users <- u
				}
			}
		}(queryParams)

		if queryParams.TotalCount >= count {
			break
		}
		queryParams = *queryParams.NextPage()
	}
	wg.Wait()
	close(users)
	fmt.Println("LEN OF CHANNEL:", len(users))
	var uss []*models.User
	fmt.Println("START READING FROM CHANNEL")
	for user := range users {
		if user == nil {
			fmt.Println("USER IS NIL")
			continue
		}
		uss = append(uss, user)
	}
	fmt.Println("USERS COUNT", len(uss))
	err = store.CreateUsers(uss)
	if err != nil {
		fmt.Println("error adding users to database:", err)
	}

	err = store.AddGroupMembers(group.ID, uss)
	if err != nil {
		log.Println("error creating users in database:", err)
	}
	fmt.Println("users added")

	groupedUsers := groupUsersByCity(uss)

	for city, users := range groupedUsers {
		err := store.BindUsersToCity(&city, users)
		if err != nil {
			log.Println("error binding cities to users:", err)
		}
	}
	return uss, nil
}

type UsersByCity map[models.UserCity][]*models.User

func groupUsersByCity(users []*models.User) UsersByCity {
	usersByCity := UsersByCity{}
	for _, u := range users {
		if u.City != nil {
			usersByCity[*u.City] = append(usersByCity[*u.City], u)
		}
	}
	return usersByCity
}

func ResolveScreenName(screenName string) (*models.Resolver, error) {
	values := make(url.Values)
	values.Set("screen_name", screenName)

	r, err := client.MakeRequest("utils.resolveScreenName", values)
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
	store.GormDB().AutoMigrate(&models.User{}, &models.Group{}, &models.UserCity{})
}
