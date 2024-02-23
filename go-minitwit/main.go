package main

import (
	"crypto/md5"
	"fmt"
	"strconv"

	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

const (
	DATABASE string = "./tmp/minitwit.db"
	PERPAGE  int    = 30
)

type User struct {
	UserID int
}

type Message struct {
	MessageID    int
	AuthorID     int
	Text         string
	PubDate      time.Time
	User         User
	Email        string
	Username     string
	Profile_link string
	Gravatar     string
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

	//adding simulatorAPI
	// registerSimulatorApi(router)

	// Start the server
	router.Run(":8081")

}

// Handlers

func userFollowActionHandler(c *gin.Context) {

	session := sessions.Default(c)

	userID, errID := c.Cookie("UserID")
	if errID != nil {
		session.AddFlash("You need to login before continuing to follow or unfollow.")
		session.Save()
		c.Redirect(http.StatusFound, "/login")
		return

	}
	profileUserName := c.Param("username")
	profileUser, err := getUserByUsername(profileUserName)
	if err != nil {
		fmt.Println("get user failed with:", err)
		c.Redirect(http.StatusFound, "/public")
		return
	}
	profileUserID := fmt.Sprintf("%v", profileUser[0]["user_id"])

	action := c.Param("action")

	if action == "/follow" {
		fmt.Println("following process triggered")
		followUser(userID, profileUserID)
		session.AddFlash("You are now following " + profileUserName)
	}
	if action == "/unfollow" {
		fmt.Println("Unfollowing process triggered")
		unfollowUser(userID, profileUserID)
		session.AddFlash("You are no longer following " + profileUserName)
	}
	session.Save()
	c.Redirect(http.StatusFound, "/"+profileUserName)
}

func publicTimelineHandler(c *gin.Context) {

	messages, err := getPublicMessages()
	if err != nil {
		return
	}
	formattedMessages := format_messages(messages)

	context := gin.H{
		"TimelineBody": true, // This seems to be a flag you use to render specific parts of your layout
		"Endpoint":     "public_timeline",
		"Messages":     formattedMessages,
	}

	userID, errID := c.Cookie("UserID")
	if errID == nil {
		context["UserID"] = userID
		userName, errName := getUserNameByUserID(userID)

		if errName == nil {
			context["UserName"] = userName
		}
	}
	// Render timeline template with the context including link variables
	c.HTML(http.StatusOK, "timeline.html", context)
}

func userTimelineHandler(c *gin.Context) {
	session := sessions.Default(c)
	flashMessages := session.Flashes()
	session.Save()
	profileUserName := c.Param("username")
	profileUser, err := getUserByUsername(profileUserName)

	if profileUser == nil {
		c.AbortWithStatus(404)
		return
	}
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// does the logged in user follow them
	followed := false
	pUserId := profileUser[0]["user_id"].(int64)
	profileName := profileUser[0]["username"]
	userID, errID := c.Cookie("UserID")
	userIDInt64, err := strconv.ParseInt(userID, 10, 64)

	userName, _ := getUserNameByUserID(userID)

	if errID == nil {
		followed = checkFollowStatus(userIDInt64, pUserId)
	}

	messages, err := getUserMessages(pUserId)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	formattedMessages := format_messages(messages)

	c.HTML(http.StatusOK, "timeline.html", gin.H{
		"TimelineBody":    true,
		"Endpoint":        "user_timeline",
		"UserID":          userIDInt64,
		"UserName":        userName,
		"Messages":        formattedMessages,
		"Followed":        followed,
		"ProfileUser":     pUserId,
		"ProfileUserName": profileName,
		"Flashes":         flashMessages,
	})
}

func myTimelineHandler(c *gin.Context) {
	userID, err := c.Cookie("UserID")

	if err != nil {
		c.Redirect(http.StatusFound, "/public")
		return
	}

	userName, err := getUserNameByUserID(userID)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	session := sessions.Default(c)
	flashMessages := session.Flashes()
	session.Save() // Clear flashes after retrieving

	messages, err := getMyMessages(userID)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	formattedMessages := format_messages(messages)

	// For template rendering with Gin
	c.HTML(http.StatusOK, "timeline.html", gin.H{
		"TimelineBody": true,
		"Endpoint":     "my_timeline",
		"UserID":       userID,
		"UserName":     userName,
		"Messages":     formattedMessages,
		"Followed":     false,
		"ProfileUser":  userID,
		"Flashes":      flashMessages,
	})
}

func addMessageHandler(c *gin.Context) {
	userID, err := c.Cookie("UserID")
	if err != nil {
		c.Redirect(http.StatusFound, "/public")
		return
	}

	var errorData string
	if c.Request.Method == http.MethodPost {
		err := c.Request.ParseForm()
		if err != nil {
			errorData = "Failed to parse form data"
			c.HTML(http.StatusBadRequest, "timeline.html", gin.H{
				"RegisterBody": true,
				"Error":        errorData,
			})
			return
		}

		// Validate form data
		text := c.Request.FormValue("text")

		if text == "" {
			errorData = "You have to enter a value"
		} else {
			err := addMessage(text, userID)
			if err != nil {
				fmt.Println("fuck my life")
				errorData = "Failed to register user"
				c.HTML(http.StatusInternalServerError, "timeline.html", gin.H{
					"RegisterBody": true,
					"Error":        errorData,
				})
				return
			}
			// Redirect to login page after successful registration
			c.Redirect(http.StatusSeeOther, "/")
			// todo: flash
			return
		}
	}
	c.HTML(http.StatusOK, "timeline.html", gin.H{
		"RegisterBody": true,
		"Error":        errorData,
	})
}

func registerHandler(c *gin.Context) {

	session := sessions.Default(c)

	userID, exists := c.Get("UserID")
	if exists {
		c.Redirect(http.StatusFound, "/"+userID.(string))
		return
	}

	var errorData string
	if c.Request.Method == http.MethodPost {
		err := c.Request.ParseForm()
		if err != nil {
			errorData = "Failed to parse form data"
			c.HTML(http.StatusBadRequest, "register.html", gin.H{
				"RegisterBody": true,
				"Error":        errorData,
			})
			return
		}

		// Validate form data
		userName := c.Request.FormValue("username")
		email := c.Request.FormValue("email")
		password := c.Request.FormValue("password")
		passwordConfirm := c.Request.FormValue("passwordConfirm")

		userID, err := getUserIDByUsername(userName)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if userName == "" {
			errorData = "You have to enter a username"
		} else if email == "" || !strings.Contains(email, "@") {
			errorData = "You have to enter a valid email address"
		} else if password == "" {
			errorData = "You have to enter a password"
		} else if password != passwordConfirm {
			errorData = "The two passwords do not match"
		} else if fmt.Sprint(userID) != "0" {
			errorData = "The username is already taken"
		} else {
			hash := md5.Sum([]byte(password))
			err := registerUser(userName, email, hash)
			if err != nil {
				errorData = "Failed to register user"
				c.HTML(http.StatusInternalServerError, "register.html", gin.H{
					"RegisterBody": true,
					"Error":        errorData,
				})
				return
			}
			// Redirect to login page after successful registration
			session.AddFlash("You were successfully registered and can login now")
			session.Save()
			c.Redirect(http.StatusSeeOther, "/login")
			// todo: flash
			return
		}
	}
	c.HTML(http.StatusOK, "register.html", gin.H{
		"RegisterBody": true,
		"Error":        errorData,
	})
}

func loginHandler(c *gin.Context) {
	session := sessions.Default(c)
	flashMessages := session.Flashes()
	session.Save()

	userID, _ := c.Cookie("UserID")
	if userID != "" {
		session.AddFlash("You were logged in")
		session.Save()
		c.Redirect(http.StatusFound, "/")
		return
	}

	var errorData string

	if c.Request.Method == http.MethodPost {

		err := c.Request.ParseForm()
		if err != nil {
			errorData = "Failed to parse form data"
			c.HTML(http.StatusBadRequest, "login.html", gin.H{
				"loginBody": true,
				"Error":     errorData,
			})
			return
		}

		userName := c.Request.FormValue("username")
		password := c.Request.FormValue("password")

		user, err := getUserByUsername(userName)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if user == nil {
			errorData = "Invalid username"
		} else if !checkPasswordHash(password, user[0]["pw_hash"].(string)) {
			errorData = "Invalid password"
		} else {
			userID, err := getUserIDByUsername(userName)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			c.SetCookie("UserID", fmt.Sprint(userID), 3600, "/", "", false, true)
			session.AddFlash("You were logged in")
			session.Save()
			c.Redirect(http.StatusFound, "/")
			return
		}

	}

	c.HTML(http.StatusOK, "login.html", gin.H{
		"LoginBody": true,
		"Error":     errorData,
		"Flashes":   flashMessages,
	})
}

func logoutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.AddFlash("You were logged out")
	session.Save()
	// Invalidate the cookie by setting its max age to -1
	// will delete the cookie <- nice stuff
	c.SetCookie("UserID", "", -1, "/", "", false, true)
	// redirect the user to the home page or login page
	c.Redirect(http.StatusFound, "/login")
}

// Helper functions
// maybe there is a better way to do this
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

func format_messages(messages []map[string]interface{}) []Message {
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

		// For strings, direct type assertion is fine
		if text, ok := m["text"].(string); ok {
			msg.Text = text
		}
		if userName, ok := m["username"].(string); ok {
			msg.Username = userName
		}
		if email, ok := m["email"].(string); ok {
			msg.Email = email
		}

		link := "/" + msg.Username
		msg.Profile_link = strings.ReplaceAll(link, " ", "%20")

		gravatarURL := gravatarURL(msg.Email, 48)
		msg.Gravatar = gravatarURL

		formattedMessages = append(formattedMessages, msg)
	}

	return formattedMessages
}
