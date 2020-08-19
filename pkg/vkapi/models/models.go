package models

import (
	"encoding/json"
	"log"
	"time"
)

type GroupMembers struct {
	Count   int     `json:"count"`
	Members []*User `json:"items"`
}

type User struct {
	ID          int        `json:"id" gorm:"primary_key" db:"id"`
	FirstName   string     `json:"first_name" db:"first_name"`
	LastName    string     `json:"last_name" db:"last_name"`
	Sex         int        `json:"sex" db:"sex"`
	Photo       string     `json:"photo" db:"photo"`
	Photo200    string     `json:"photo_200" db:"photo200"`
	City        *UserCity  `json:"city"`
	CityID      int        `json:"city_id"`
	Deactivated *string    `json:"deactivated" db:"deactivated"`
	LastSeen    *time.Time `json:"last_seen" db:"last_seen"`
	Groups      []*Group   `json:"groups" gorm:"many2many:group_members"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at" sql:"DEFAULT:CURRENT_TIMESTAMP"`
	//LastSeen vkapi.LastSeen `json:"last_seen" db:"last_seen"`
}

func (u *User) UnmarshalJSON(data []byte) error {
	type Alias User
	uuu := &struct {
		*Alias
		LastSeen *LastSeen `json:"last_seen"`
	}{
		Alias: (*Alias)(u),
	}

	err := json.Unmarshal(data, uuu)
	if err != nil {
		log.Println(`error unmarshal`, err)
	}
	if uuu.LastSeen == nil {
		return nil
	}
	pTime := time.Unix(uuu.LastSeen.Time, 0)

	u.LastSeen = &pTime

	return nil
}

type LastSeen struct {
	Time     int64 `json:"time"`
	Platform int   `json:"platform"`
}

type Group struct {
	ID          int        `json:"id" gorm:"primary_key" db:"id"`
	Name        string     `json:"name" db:"name"`
	ScreenName  string     `json:"screen_name" db:"screen_name"`
	Description string     `json:"description"db:"description"`
	Activity    string     `json:"activity" db:"activity"`
	IsClosed    int        `json:"is_closed" db:"is_closed"`
	Type        string     `json:"type" db:"type"`
	Members     []*User    `json:"members" gorm:"many2many:group_members"`
	LastUpdate  *time.Time `json:"last_update" db:"last_update"`
}

type Resolver struct {
	Type     string `json:"type"`
	ObjectID int    `json:"object_id"`
}

type GroupMember struct {
	UserID  int `json:"user_id" db:"user_id"`
	GroupID int `json:"group_id" db:"group_id"`
}

type UserCity struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}
