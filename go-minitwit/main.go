package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"reflect"

	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
)

// todo: can we move these as well?
const (
	DATABASE string = "./tmp/minitwit.db"
	PERPAGE  int    = 30
)

type FilteredMsg struct {
	Content string `json:"content"`
	PubDate int64  `json:"pub_date"`
	User    string `json:"user"`
}

var dbNew *gorm.DB

func main() {

	// Using db connection (1)
	var err error
	dbNew, err = connect_DB(DATABASE)
	if err != nil {
		panic("failed to connect to database")
	}

	// Create a Gin router and set the parsed templates
	router := gin.Default()
	router.LoadHTMLGlob("./templates/*.html")

	// sessions, for cookies
	store := cookie.NewStore([]byte("devops"))
	router.Use(sessions.Sessions("session", store))

	// Static (styling)
	router.Static("/static", "./static")

	// Define routes -> Here is where the links are being registered! Check the html layout file
	// user routes
	router.GET("/", myTimelineHandler)
	router.GET("/public", publicTimelineHandler)
	router.GET("/:username", userTimelineHandler)
	router.GET("/register", registerHandler)
	router.GET("/login", loginHandler)
	router.GET("/logout", logoutHandler)
	router.GET("/:username/*action", userFollowActionHandler)

	router.POST("/register", registerHandler)
	router.POST("/login", loginHandler)
	router.POST("/add_message", addMessageHandler)

	// API routes
	// is it easier to separate the next two routes into two handlers?
	router.GET("/api/msgs", apiMsgsHandler)
	router.GET("/api/msgs/:username", apiMsgsPerUserHandler)
	router.GET("/api/fllws/:username", apiFllwsHandler)

	router.POST("/api/register", apiRegisterHandler)
	router.POST("/api/msgs/:username", apiMsgsPerUserHandler)
	router.POST("/api/fllws/:username", apiFllwsHandler)

	// some helper method to "cache" what was the latest simulator action
	router.GET("/api/latest", getLatest)

	// adding simulatorAPI
	// registerSimulatorApi(router)

	// Start the server
	router.Run(":8081")

}

// Helper functions
// todo: move these to a "helperLibrary"
func bytesToString(bytes [16]byte) string {
	var strBuilder strings.Builder

	strBuilder.WriteString("{")
	for i, b := range bytes {
		if i > 0 {
			strBuilder.WriteString(",")
		}
		strBuilder.WriteString(fmt.Sprintf("%d", b))
	}
	strBuilder.WriteString("}")

	return strBuilder.String()
}

func checkPasswordHash(userEnteredPwd string, dbpwd string) bool {
	hash := md5.Sum([]byte(userEnteredPwd))
	str := hex.EncodeToString(hash[:])
	return str == dbpwd
}

func gravatarURL(email string, size int) string {
	if size <= 0 {
		size = 80 // Default size
	}

	email = strings.ToLower(strings.TrimSpace(email))
	hash := md5.Sum([]byte(email))
	return fmt.Sprintf("http://www.gravatar.com/avatar/%x?d=identicon&s=%d", hash, size)
}

func formatMessages(messages []MessageUser) []MessageUI {
	var formattedMessages []MessageUI

	/*
		if reflect.TypeOf(m.Text).Kind() == reflect.String {
			filteredMsg.Content = m.Text
		}
	*/

	for _, m := range messages {
		var msg MessageUI
		// Use type assertion for int64, then convert to int
		if reflect.TypeOf(m.MessageID).Kind() == reflect.Int {
			msg.MessageID = int(m.MessageID)
		}
		if reflect.TypeOf(m.AuthorID).Kind() == reflect.Int {
			msg.AuthorID = int(m.AuthorID)
		}
		if reflect.TypeOf(m.UserID).Kind() == reflect.Int {
			msg.User.UserID = int(m.UserID)
		}
		if reflect.TypeOf(m.Text).Kind() == reflect.String {
			msg.Text = m.Text
		}
		if reflect.TypeOf(m.Username).Kind() == reflect.String {
			msg.Username = m.Username
		}
		if reflect.TypeOf(m.Email).Kind() == reflect.String {
			msg.Email = m.Email
		}
		if reflect.TypeOf(m.PubDate).Kind() == reflect.Int {
			pubDateTime := time.Unix(int64(m.PubDate), 0)
			msg.PubDate = pubDateTime.Format("02/01/2006 15:04:05") // go time layout format is weird 1,2,3,4,5,6 ¬¬
		}
		link := "/" + msg.Username
		msg.Profile_link = strings.ReplaceAll(link, " ", "%20")

		gravatarURL := gravatarURL(msg.Email, 48)
		msg.Gravatar = gravatarURL

		formattedMessages = append(formattedMessages, msg)
	}

	return formattedMessages
}

func filterMessages(messages []MessageUser) []FilteredMsg {
	var filteredMessages []FilteredMsg
	for _, m := range messages {
		var filteredMsg FilteredMsg
		// content
		if reflect.TypeOf(m.Text).Kind() == reflect.String {
			filteredMsg.Content = m.Text
		}

		// publication date
		filteredMsg.PubDate = int64(m.PubDate)

		// user
		if reflect.TypeOf(m.Username).Kind() == reflect.String {
			filteredMsg.User = m.Username
		}

		filteredMessages = append(filteredMessages, filteredMsg)
	}
	return filteredMessages
}
