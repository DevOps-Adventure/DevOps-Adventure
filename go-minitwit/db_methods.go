package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import the SQLite3 driver
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

/*
	MODELS FOR GORM
*/

type User struct {
	UserID   int `gorm:"primaryKey"`
	Username string
	Email    string
	PwHash   string
}

type Message struct {
	MessageID int `gorm:"primaryKey"`
	AuthorID  int
	Text      string
	PubDate   int
	Flagged   int
}

type MessageUser struct {
	MessageID int `gorm:"primaryKey"`
	AuthorID  int
	Text      string
	PubDate   int
	Flagged   int
	UserID    int `gorm:"primaryKey"`
	Username  string
	Email     string
	PwHash    string
}

type MessageUI struct {
	MessageID    int
	AuthorID     int
	Text         string
	PubDate      string
	Flagged      bool
	User         User
	Email        string
	Username     string
	Profile_link string
	Gravatar     string
}

type Follower struct {
	WhoID  int
	WhomID int
}

// declare table names so that GORM doesn't fall on default values
func (User) TableName() string {
	return "user"
}

func (Message) TableName() string {
	return "message"
}

func (Follower) TableName() string {
	return "follower"
}

/*
	CONNECT, INIT AND QUERY DB
*/

func connect_dev_DB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}})
	if err != nil {
		panic("failed to connect to database")
	}

	db.AutoMigrate(&User{}, &Message{}, &Follower{})

	return db, nil
}

func connect_prod_DB() (*gorm.DB, error) {
	dsn := os.Getenv("DBUSER") + ":" + os.Getenv("DBPASS") + "@tcp(db-mysql-fra1-34588-do-user-15917069-0.c.db.ondigitalocean.com:25060)/devopsadventure"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("gorm Db connection ", err)
		return nil, err
	}

	return db, nil
}

/*
	GET DATA
*/

// fetches all public messages for display.
func getPublicMessages(numMsgs int) ([]MessageUser, error) {
	var messages []MessageUser
	dbNew.Table("message").
		Select("message.*, user.*").
		Joins("JOIN user AS user ON message.author_id = user.user_id").
		Where("message.flagged = ?", 0).
		Order("message.pub_date DESC").
		Limit(numMsgs).
		Find(&messages)

	if dbNew.Error != nil {
		log.Fatal(dbNew.Error)
		return nil, dbNew.Error
	}

	return messages, nil
}

// fetches all messages from picked user
func getUserMessages(pUserId int, numMsgs int) ([]MessageUser, error) {

	var messages []MessageUser
	dbNew.Table("message").
		Select("message.*, user.*").
		Joins("JOIN user ON user.user_id = message.author_id").
		Where("user.user_id = ?", pUserId).
		Order("message.pub_date desc").
		Limit(numMsgs).
		Find(&messages)

	if dbNew.Error != nil {
		log.Fatal(dbNew.Error)
		return nil, dbNew.Error
	}

	return messages, nil
}

// check whether the given user is followed by logged in
func checkFollowStatus(userID int, pUserID int) (bool, error) {

	if userID == pUserID {
		return false, nil
	}

	var follower Follower
	dbNew.Where("who_id = ? AND whom_id = ?", userID, pUserID).First(&follower)

	if dbNew.Error != nil {
		log.Fatal(dbNew.Error)
		return false, dbNew.Error
	}

	if follower.WhoID == 0 || follower.WhomID == 0 {
		return false, nil
	}

	return true, nil
}

// fetches all messages for the current logged in user for 'My Timeline'
func getMyMessages(userID string) ([]MessageUser, error) {
	var messages []MessageUser

	subQuery := dbNew.Table("follower").
		Select("whom_id").
		Where("who_id = ?", userID)

	var followerIDs []int

	// Find the IDs from the subquery
	if err := subQuery.Find(&followerIDs).Error; err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Use the retrieved followerIDs in the main query
	dbNew.Table("message").
		Select("message.*, user.*").
		Joins("JOIN user ON message.author_id = user.user_id").
		Where("message.flagged = ? AND (user.user_id = ? OR user.user_id IN (?))", 0, userID, followerIDs).
		Order("message.pub_date DESC").
		Find(&messages)

	if dbNew.Error != nil {
		log.Fatal(dbNew.Error)
		return nil, dbNew.Error
	}

	return messages, nil
}

// fetches a user by their ID
func getUserIDByUsername(userName string) (int, error) {
	var user User
	dbNew.Where("username = ?", userName).First(&user)

	if user.UserID == 0 {
		return -1, nil
	} else {
		return user.UserID, nil
	}
}

// fetches a username by their ID
func getUserNameByUserID(userID string) (string, error) {
	var user User
	dbNew.First(&user, userID)

	if dbNew.Error != nil {
		log.Fatal(dbNew.Error)
		return "", dbNew.Error
	}

	return user.Username, nil
}

func getUserByUsername(userName string) (User, error) {
	var user User
	dbNew.Where("username = ?", userName).First(&user)

	if dbNew.Error != nil {
		log.Fatal(dbNew.Error)
		return user, dbNew.Error
	}

	return user, nil

}

/*
	POST DATA
*/

// registers a new user
func registerUser(userName string, email string, password [16]byte) error {
	pwHashString := hex.EncodeToString(password[:])

	newUser := User{
		Username: userName,
		Email:    email,
		PwHash:   pwHashString,
	}

	dbNew.Table("user").
		Create(&newUser)

	if dbNew.Error != nil {
		log.Fatal(dbNew.Error)
		return dbNew.Error
	}

	return nil
}

// adds a new message to the database
func addMessage(text string, author_id int) error {

	currentTime := time.Now().UTC()
	unixTimestamp := currentTime.Unix()

	newMessage := Message{
		AuthorID: author_id,
		Text:     text,
		PubDate:  int(unixTimestamp),
		Flagged:  0, // Default to false for flagged
	}

	dbNew.Table("message").
		Create(&newMessage)

	if dbNew.Error != nil {
		log.Fatal(dbNew.Error)
		return dbNew.Error
	}

	return nil

}

// followUser adds a new follower to the database
func followUser(userID string, profileUserID string) error {
	userIDInt, errz := strconv.Atoi(userID)
	profileUserIDInt, errx := strconv.Atoi(profileUserID)

	if errz != nil {
		log.Fatal(errz)
		return errz
	} else if errx != nil {
		log.Fatal(errx)
		return errx
	}

	// following relationship already exists
	var count int64
	dbNew.Table("follower").
		Model(&Follower{}).
		Where("who_id = ? AND whom_id = ?", userIDInt, profileUserIDInt).
		Count(&count)
	if count > 0 {
		return nil
	}

	newFollower := Follower{
		WhoID:  userIDInt,
		WhomID: profileUserIDInt,
	}

	dbNew.Table("follower").
		Create(&newFollower)

	if dbNew.Error != nil {
		log.Fatal(dbNew.Error)
		return dbNew.Error
	}

	return nil
}

// unfollowUser removes a follower from the database
func unfollowUser(userID string, profileUserID string) error {
	userIDInt, errz := strconv.Atoi(userID)
	profileUserIDInt, errx := strconv.Atoi(profileUserID)

	if errz != nil {
		log.Fatal(errz)
		return errz
	} else if errx != nil {
		log.Fatal(errx)
		return errx
	}

	dbNew.Where("who_id = ? AND whom_id = ?", userIDInt, profileUserIDInt).Delete(&Follower{})

	if dbNew.Error != nil {
		log.Fatal(dbNew.Error)
		return dbNew.Error
	}

	return nil
}

// getFollowers fetches up to `limit` followers for the user identified by userID
func getFollowers(userID string, limit int) ([]User, error) {

	var users []User

	dbNew.
		Select("user.*").
		Joins("INNER JOIN follower ON user.user_id = follower.who_id").
		Where("follower.whom_id = ?", userID).
		Limit(limit).
		Find(&users)

	if dbNew.Error != nil {
		log.Fatal(dbNew.Error)
		return users, dbNew.Error
	}
	return users, nil
}

// getFollowing fetches up to `limit` users that the user identified by userID is following
func getFollowing(userID string, limit int) ([]User, error) {
	var users []User

	dbNew.
		Select("user.*").
		Joins("INNER JOIN follower ON user.user_id = follower.whom_id").
		Where("follower.who_id = ?", userID).
		Limit(limit).
		Find(&users)

	if dbNew.Error != nil {
		log.Fatal(dbNew.Error)
		return users, dbNew.Error
	}

	return users, nil
}
