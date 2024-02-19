package main

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
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
	//registerSimulatorApi(router)

	// Start the server
	router.Run(":8080")

}

func connect_db(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// db initialization (2)
// to keep separate functions of connection and initialization of the db. (here the db is structured with specific schema/format)
func init_db(db *sql.DB, schemaFile string) error {
	schema, err := os.ReadFile(schemaFile)
	if err != nil {
		return err
	}

	// Executing the schema SQL after it is being read in the previous step
	_, err = db.Exec(string(schema))
	if err != nil {
		return err
	}
	return err
}

// db query that returns list of dictionaries (3)
func query_db(db *sql.DB, query string, args []interface{}, one bool) ([]map[string]interface{}, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columns {
			columnPointers[i] = &values[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]interface{})
		for i, colName := range columns {
			val := columnPointers[i].(*interface{})
			rowMap[colName] = *val
		}

		result = append(result, rowMap)
		if one {
			break
		}
	}
	return result, nil
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
	profileUser, err := getUserByUsername2(profileUserName)
	if err != nil {
		fmt.Println("get user failed with:", err)
		c.Redirect(http.StatusFound, "/public")
		return
	}
	profileUserID := fmt.Sprintf("%v", profileUser[0]["user_id"])

	action := c.Param("action")

	if action == "/follow" {
		fmt.Println("following process triggered")
		followUser2(userID, profileUserID)
		session.AddFlash("You are now following " + profileUserName)
	}
	if action == "/unfollow" {
		fmt.Println("Unfollowing process triggered")
		unfollowUser2(userID, profileUserID)
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
		userName, errName := getUserNameByUserID2(userID)

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
	profileUser, err := getUserByUsername2(profileUserName)

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

	userName, _ := getUserNameByUserID2(userID)

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

	userName, err := getUserNameByUserID2(userID)

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
			err := addMessage2(text, userID)
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
		password2 := c.Request.FormValue("password2")

		userID, err := getUserIDByUsername2(userName)
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
		} else if password != password2 {
			errorData = "The two passwords do not match"
		} else if fmt.Sprint(userID) != "0" {
			errorData = "The username is already taken"
		} else {
			hash := md5.Sum([]byte(password))
			err := registerUser2(userName, email, hash)
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

		user, err := getUserByUsername2(userName)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if user == nil {
			errorData = "Invalid username"
		} else if !checkPasswordHash(password, user[0]["pw_hash"].(string)) {
			errorData = "Invalid password"
		} else {
			userID, err := getUserIDByUsername2(userName)
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

func getUserIDByUsername(userName string) (int64, error) {
	var db, err = connect_db(DATABASE)
	if err != nil {
		return 0, err
	}

	query := `select * from user where username = ?`
	args := []interface{}{userName}
	profile_user, err := query_db(db, query, args, false)

	if profile_user == nil {
		return 0, err
	}

	return profile_user[0]["user_id"].(int64), err
}

func getUserNameByUserID(userID string) (string, error) {
	var db, err = connect_db(DATABASE)
	if err != nil {
		return "", err
	}

	query := `select * from user where user_id = ?`
	args := []interface{}{userID}
	profile_user, err := query_db(db, query, args, false)

	if profile_user == nil {
		return "no name", err
	}
	return profile_user[0]["username"].(string), err
}

func getUserByUsername(userName string) ([]map[string]interface{}, error) {
	var db, err = connect_db(DATABASE)
	if err != nil {
		return nil, err
	}

	query := `select * from user where username = ?`
	args := []interface{}{userName}
	profile_user, err := query_db(db, query, args, false)

	if profile_user == nil {
		return nil, err
	}

	return profile_user, err
}

func registerUser(userName string, email string, password [16]byte) error {
	query := `insert into user (username, email, pw_hash) values (?, ?, ?)`
	var db, err = connect_db(DATABASE)
	if err != nil {
		return err
	}
	args := []interface{}{userName, email, pq.Array(password)}
	messages, err := query_db(db, query, args, false)
	fmt.Println("this is the messages", messages)
	return err
}

func followUser(userID string, profileUserID string) error {
	query := `insert into follower (who_id, whom_id) values (?, ?)`
	var db, err = connect_db(DATABASE)
	if err != nil {
		return err
	}
	args := []interface{}{userID, profileUserID}
	messages, err := query_db(db, query, args, false)
	fmt.Println(messages)
	return err
}

func unfollowUser(userID string, profileUserID string) error {
	query := `delete from follower where who_id=? and whom_id=?`
	var db, err = connect_db(DATABASE)
	if err != nil {
		return err
	}
	args := []interface{}{userID, profileUserID}
	messages, err := query_db(db, query, args, false)
	fmt.Println(messages)
	return err
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
