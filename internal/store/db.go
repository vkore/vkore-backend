package store

import (
	"database/sql"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/jmoiron/sqlx"
	"github.com/vkore/vkore/pkg/vkapi/models"
	"log"
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
		if user.Deactivated != nil {
			continue
		}
		uss = append(uss, user)
	}

	insertUsers := func(args interface{}) (sql.Result, error) {
		return db.NamedExec("INSERT INTO users (id, first_name, last_name, sex, photo, deactivated, last_seen) VALUES (:id, :first_name, :last_name, :sex, :photo, :deactivated, :last_seen) ON CONFLICT DO NOTHING", args)
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
		return db.NamedExec("INSERT INTO group_members (user_id, group_id) VALUES (:user_id, :group_id)", args)
	}

	var err error
	usersCount := len(members)
	for i := 0; i < usersCount/maxRecordsPerQuery; i++ {
		usersToWriteCount := groupMembers[:maxRecordsPerQuery]
		_, err = insertUsers(usersToWriteCount)
		if err != nil {
			log.Println(`error creating user:`, err)
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

func GetUsers(filters ...*Filter) []*models.User {
	var users []*models.User

	query := gormDB
	for _, filter := range filters {
		query = query.Where(filter.Query, filter.Args...)
	}
	query.Find(&users)
	return users
}

func GetGroupLastUpdate(groupID int) (*time.Time, error) {
	var group models.Group

	if err := gormDB.Where("id = ?", groupID).First(&group).Error; err != nil {
		return nil, fmt.Errorf(`can't get gruop "%v": %v`, groupID, err)
	}
	return group.LastUpdate, nil
}