// api.go
/*
Example of API request
c.Request.Method gives you the HTTP method of the request (GET, POST, etc.).
c.Request.URL gives you the URL of the request.
c.Request.Header gives you the request headers.
c.Request.Body gives you the request body, which you can parse according to the content type (JSON, form, etc.).
*/

package main

import (
	"crypto/md5"

	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

const (
	DATABASE2 string = "./tmp/minitwit.db"
	PERPAGE2  int    = 30
)

func mainAPI() {

	// Connect to the database
	db, err := connect_DB2(DATABASE)
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
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

	// Define routes
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

	//adding the simulatorApi
	registerSimulatorApi(router)

	// Start the server
	router.Run(":8081")
}

// Define your route handlers here
func myTimelineHandlerAPI(c *gin.Context) {
	// Not implemented
	if IsSimulatorRequest(c) {
	} else {
		messages, err := getPublicMessages() // Assume this function exists in db_connection.go
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		formattedMessages := format_messages(messages)
		c.HTML(http.StatusOK, "timeline.html", gin.H{
			"TimelineBody": true,
			"Endpoint":     "public_timeline",
			"Messages":     formattedMessages,
		})
	}
}

func publicTimelineHandlerAPI(c *gin.Context) {
	if IsSimulatorRequest(c) {

	} else {
		messages, err := getPublicMessages() // Assume this function exists in db_connection.go
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		formattedMessages := format_messages(messages) // format_messages needs to be implemented
		c.HTML(http.StatusOK, "timeline.html", gin.H{
			"TimelineBody": true,
			"Endpoint":     "public_timeline",
			"Messages":     formattedMessages,
		})
	}
}

func userTimelineHandlerAPI(c *gin.Context) {
	// Not implemented
	if IsSimulatorRequest(c) {
	} else {
		username := c.Param("username")
		profileUser, err := getUserByUsername2(username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		if len(profileUser) == 0 {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		pUserID := profileUser[0]["user_id"].(int64)

		var followed bool
		if userIDValue, exists := c.Get("userID"); exists {
			// Safely assert userIDValue to int64
			userID, ok := userIDValue.(int64)
			if !ok {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			followed = checkFollowStatus(userID, pUserID)

		}

		messages, err := getUserMessages(pUserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		formattedMessages := format_messages(messages) // Assumes format_messages is defined
		c.HTML(http.StatusOK, "timeline.html", gin.H{
			"TimelineBody": true,
			"Endpoint":     "user_timeline",
			"Username":     username,
			"Messages":     formattedMessages,
			"Followed":     followed,
		})
	}
}

func registerHandlerAPI(c *gin.Context) (string, int) {

	// Parse JSON request body
	// better to use a struct. SH: Did not test it yet!
	var requestJSON struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&requestJSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return "Invalid JSON format", http.StatusBadRequest
	}

	// Validate request data
	if requestJSON.Username == "" || requestJSON.Email == "" || requestJSON.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return "Missing required fields", http.StatusBadRequest
	}

	// Check if username is already taken
	exists, err := userExists(requestJSON.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check username availability"})
		return "Failed to check username availability", http.StatusInternalServerError
	}
	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The username is already taken"})
		return "The username is already taken", http.StatusBadRequest
	}

	// Hash the password
	hashedPassword := md5.Sum([]byte(requestJSON.Password))

	// Register the user
	if err := registerUser2(requestJSON.Username, requestJSON.Email, hashedPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return "Failed to register user", http.StatusInternalServerError
	}

	// Return success
	return "", http.StatusNoContent
}

func loginHandlerAPI(c *gin.Context) {
	// Not implemented
}

func logoutHandlerAPI(c *gin.Context) {
	// Not implemented
}

func userFollowActionHandlerAPI(c *gin.Context) {
	// Not implemented
}

func addMessageHandlerAPI(c *gin.Context) {
	// Not implemented
}
