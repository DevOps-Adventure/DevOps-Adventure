package main

import (
	"crypto/md5"
	"fmt"

	"log"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// todo: can we move these as well?
const (
	DATABASE string = "./tmp/minitwit_empty.db"
	PERPAGE  int    = 30
)

// todo: can we move these?
type User struct {
	UserID int
}

type Message struct {
	MessageID    int
	AuthorID     int
	Text         string
	PubDate      string
	User         User
	Email        string
	Username     string
	Profile_link string
	Gravatar     string
}

type FilteredMsg struct {
	Content string `json:"content"`
	PubDate int64  `json:"pub_date"`
	User    string `json:"user"`
}

func main() {

	//using db connection (1)
	db, err := connect_db(DATABASE)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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
	bytes := md5.Sum([]byte(userEnteredPwd))
	str := bytesToString(bytes)
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

func formatMessages(messages []map[string]interface{}) []Message {
	var formattedMessages []Message
	for _, m := range messages {
		var msg Message
		// Use type assertion for int64, then convert to int
		if id, ok := m["message_id"].(int64); ok {
			msg.MessageID = int(id)
		}
		if authorID, ok := m["author_id"].(int64); ok {
			msg.AuthorID = int(authorID)
		}
		if userID, ok := m["user_id"].(int64); ok {
			msg.User.UserID = int(userID)
		}
		if text, ok := m["text"].(string); ok {
			msg.Text = text
		}
		if userName, ok := m["username"].(string); ok {
			msg.Username = userName
		}
		if email, ok := m["email"].(string); ok {
			msg.Email = email
		}
		if pubDate, ok := m["pub_date"].(int64); ok {
			pubDateTime := time.Unix(pubDate, 0)
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

func filterMessages(messages []map[string]interface{}) []FilteredMsg {
	var filteredMessages []FilteredMsg
	for _, m := range messages {
		var filteredMsg FilteredMsg
		// content
		if text, ok := m["text"].(string); ok {
			filteredMsg.Content = text
		}

		// publication date
		if pubDate, ok := m["pub_date"].(int64); ok {
			filteredMsg.PubDate = pubDate
		}

		// user
		if userName, ok := m["username"].(string); ok {
			filteredMsg.User = userName
		}

		filteredMessages = append(filteredMessages, filteredMsg)
	}
	return filteredMessages
}
