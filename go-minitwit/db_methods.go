package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import the SQLite3 driver
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

/*
	CONNECT, INIT AND QUERY DB
*/

func connect_DB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}})
	if err != nil {
		panic("failed to connect to database")
	}

	db.AutoMigrate(&User{}, &Message{}, &Follower{})

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
func getUserMessages(pUserId int64, numMsgs int) ([]MessageUser, error) {

	var messages []MessageUser
	fmt.Println("userid:", pUserId)
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
func checkFollowStatus(userID int64, pUserID int64) (bool, error) {

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
func getUserIDByUsername(userName string) (int64, error) {
	var user User
	dbNew.Where("username = ?", userName).First(&user)

	if user.UserID == 0 {
		return -1, nil
	} else {
		return int64(user.UserID), nil
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

	dbNew.Create(&newUser)

	if dbNew.Error != nil {
		log.Fatal(dbNew.Error)
		return dbNew.Error
	}

	return nil

}

// adds a new message to the database
func addMessage(text string, author_id string) error {

	authorIDintValue, err := strconv.Atoi(author_id)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	currentTime := time.Now().UTC()
	unixTimestamp := currentTime.Unix()

	newMessage := Message{
		AuthorID: authorIDintValue,
		Text:     text,
		PubDate:  int(unixTimestamp),
		Flagged:  0, // Default to false for flagged
	}

	dbNew.Create(&newMessage)

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

	newFollower := Follower{
		WhoID:  userIDInt,
		WhomID: profileUserIDInt,
	}

	dbNew.Create(&newFollower)

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

/*
// userExists checks if a user with the given username already exists in the database.
func userExists(username string) (bool, error) {
	query := "SELECT COUNT(*) FROM user WHERE username = ?"
	db, err := connect_db(DATABASE)
	if err != nil {
		return false, err
	}
	defer db.Close()

	var count int
	err = db.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, err
	}

	// 	// If count is greater than 0, the user exists
	// 	return count > 0, nil
	// }
	return true, nil
}
*/
