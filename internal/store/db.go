package store

import (
	"database/sql"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/jmoiron/sqlx"
	"github.com/vkore/vkore/pkg/vkapi/models"
	"log"
	"strconv"
	"strings"
	"time"
)

var gormDB *gorm.DB
var db *sqlx.DB

// Init - creates new gorm database connection instance
func Init() {
	var err error
	gormDB, err = gorm.Open("sqlite3", "vkore.db")
	if err != nil {
		log.Fatalln("Can't open database connection:", err)
	}
	db, err = sqlx.Open("sqlite3", "vkore.db")
	if err != nil {
		log.Fatalln("Can't open database connection:", err)
	}
}

func GormDB() *gorm.DB {
	return gormDB
}

func CreateGroup(group *models.Group) error {
	return gormDB.Create(group).Error
}

//noinspection GoNilness
func CreateUsers(users []*models.User) error {
	const maxRecordsPerQuery = 100
	if len(users) == 0 {
		return nil
	}
	var uss []interface{}

	for _, user := range users {
		if user == nil || user.Deactivated != nil {
			continue
		}
		uss = append(uss, user)
	}
	insertUsers := func(args interface{}) (sql.Result, error) {
		return db.NamedExec("INSERT INTO users (id, first_name, last_name, sex, photo, photo200, deactivated, last_seen, status) VALUES (:id, :first_name, :last_name, :sex, :photo, :photo200, :deactivated, :last_seen, :status) ON CONFLICT DO NOTHING", args)
	}

	var err error
	usersCount := len(uss)
	for i := 0; i < usersCount/maxRecordsPerQuery; i++ {
		usersToWriteCount := uss[:maxRecordsPerQuery]
		_, err = insertUsers(usersToWriteCount)
		if err != nil {
			log.Println(`error creating user:`, err)
		}
		uss = uss[maxRecordsPerQuery:]
	}
	_, err = insertUsers(uss)

	return err
}

func AddGroupMembers(groupID int, members []*models.User) error {
	if len(members) == 0 {
		return nil
	}
	const maxRecordsPerQuery = 450
	groupMembers := make([]*models.GroupMember, len(members))

	for i, member := range members {
		groupMembers[i] = &models.GroupMember{
			UserID:  member.ID,
			GroupID: groupID,
		}
	}

	insertUsers := func(args interface{}) (sql.Result, error) {
		return db.NamedExec("INSERT INTO group_members (user_id, group_id) VALUES (:user_id, :group_id) ON CONFLICT DO NOTHING", args)
	}

	var err error
	usersCount := len(members)
	for i := 0; i < usersCount/maxRecordsPerQuery; i++ {
		usersToWriteCount := groupMembers[:maxRecordsPerQuery]
		_, err = insertUsers(usersToWriteCount)
		if err != nil {
			log.Println(`error creating user relation:`, err)
		}
		groupMembers = groupMembers[maxRecordsPerQuery:]
	}
	_, err = db.Exec("UPDATE groups SET last_update = ? WHERE id = ?;", time.Now(), groupID)
	if err != nil {
		return fmt.Errorf(`error updating group: %v`, err)
	}

	log.Println("group members added")
	return nil
}

type Filter struct {
	Query interface{}
	Args  []interface{}
}

func GetUsers(limit, offset int, filters ...*Filter) ([]*models.User, int) {
	var users []*models.User

	query := gormDB
	for _, filter := range filters {
		query = query.Where(filter.Query, filter.Args...)
	}
	var count int
	query.Model(&models.User{}).Count(&count)
	query.Limit(limit).Offset(offset).Find(&users)
	return users, count
}

func GetGroupLastUpdate(groupID int) (*time.Time, error) {
	var group models.Group
	if err := gormDB.Where("id = ?", groupID).First(&group).Error; err != nil {
		return nil, fmt.Errorf(`can't get gruop "%v": %v`, groupID, err)
	}
	return group.LastUpdate, nil
}

func GetGroupByScreenName(name string) (group *models.Group, err error) {
	if err = gormDB.Where("screen_name = ?", name).First(&group).Error; gorm.IsRecordNotFoundError(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf(`can't get gruop "%v": %v`, name, err)
	}
	return
}

func GetAllGroups() (groups []*models.Group, err error) {
	err = gormDB.Find(&groups).Error
	return
}

func BindUsersToCity(city *models.UserCity, users []*models.User) error {
	if err := gormDB.Where(models.UserCity{ID: city.ID}).Attrs(city).FirstOrCreate(city).Error; err != nil {
		log.Printf("error get or create gorup in database: %v", err)
		return err
	}

	gormDB.Where([]int64{20, 21, 22}).UpdateColumn()

	var usersIDs []string
	for _, user := range users {
		usersIDs = append(usersIDs, strconv.Itoa(user.ID))
	}

	_, err := db.Exec(`UPDATE users SET city_id = ? WHERE id IN (?);`, city.ID, strings.Join(usersIDs, ","))

	return err
}

func GetAllCities() (cities []*models.UserCity, err error) {
	err = gormDB.Find(&cities).Error
	return
}
