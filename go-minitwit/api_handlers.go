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
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// struct for error data (need to be JSON before return)
type ErrorData struct {
	status    int
	error_msg string
}

func updateLatest() {
	return
}

func getLatest(c *gin.Context) {
	return
}

/*
POST
Takes data from the POST and registers a user in the db
returns: ("", 204) or ({"status": 404, "error_msg": error}, 404)
*/
func apiRegisterHandler(c *gin.Context) {
	fmt.Println("apiRegisterHandler!")
	userID, exists := c.Get("UserID")
	fmt.Println("userID:", userID)
	if exists {
		fmt.Println("userID:", userID)
		return
	}

	errorData := ErrorData{
		status:    0,
		error_msg: "",
	}

	if c.Request.Method == http.MethodPost {
		err := c.Request.ParseForm()
		if err != nil {
			errorData.status = http.StatusBadRequest
			errorData.error_msg = "Could not parse the request"
			// we need to crash the program here, otherwise the rest of the method may fail
		}

		// Validate form data
		userName := c.Request.FormValue("username")
		email := c.Request.FormValue("email")
		password := c.Request.FormValue("password")
		passwordConfirm := c.Request.FormValue("passwordConfirm")

		//Get user ID
		userID, err := getUserIDByUsername(userName)
		if err != nil {
			errorData.status = http.StatusBadRequest
			errorData.error_msg = "Failed to get userID"
			// we need to crash the program here, otherwise the rest of the method may fail
		}

		if userName == "" {
			errorData.status = http.StatusBadRequest
			errorData.error_msg = "You have to enter a username"
		} else if email == "" || !strings.Contains(email, "@") {
			errorData.status = http.StatusBadRequest
			errorData.error_msg = "You have to enter a valid email address"
		} else if password == "" {
			errorData.status = http.StatusBadRequest
			errorData.error_msg = "You have to enter a password"
		} else if password != passwordConfirm {
			errorData.status = http.StatusBadRequest
			errorData.error_msg = "The two passwords do not match"
		} else if fmt.Sprint(userID) != "-1" {
			errorData.status = http.StatusBadRequest
			errorData.error_msg = "The username is already taken"
		} else {
			hash := md5.Sum([]byte(password))
			err := registerUser(userName, email, hash)
			if err != nil {
				errorData.status = http.StatusBadRequest
				errorData.error_msg = "Failed to register user"
			}
		}

		// we need to crash the program here, otherwise the rest of the method may fail
	}

	if errorData.error_msg != "" {
		c.JSON(errorData.status, errorData)
		return
	}

	c.String(http.StatusOK, "")
}

/*
GET and POST
if GET:

	if :username is defined:
		return: all messages by that user, status code
	else:
		return: all messages in db, status code

else if POST:

	write message to db
	return: status code
*/
type FilteredMsg struct {
	content  string
	pub_date string
	user     string
}

func apiMsgsHandler(c *gin.Context) {
	// todo: update this request to be the latest
	// todo: check if this is not from sim response
	var filteredMessages []FilteredMsg

	errorData := ErrorData{
		status:    0,
		error_msg: "",
	}

	numMsgs := c.Request.Header.Get("no")
	numMsgsInt, err := strconv.Atoi(numMsgs)
	// fallback on default value
	if err != nil {
		numMsgsInt = 100
	}

	messages, err := getPublicMessages(numMsgsInt)
	if err != nil {
		errorData.status = http.StatusBadRequest
		errorData.error_msg = "Failed to get userID"
		c.AbortWithStatusJSON(404, errorData)
	}

	for _, m := range messages {
		var filteredMsg FilteredMsg

		// content
		if text, ok := m["text"].(string); ok {
			filteredMsg.content = text
		}

		// publication date
		if pubDate, ok := m["pub_date"].(string); ok {
			filteredMsg.pub_date = pubDate
		}

		// user
		if userName, ok := m["username"].(string); ok {
			filteredMsg.user = userName
		}

		filteredMessages = append(filteredMessages, filteredMsg)
	}

	c.JSON(http.StatusOK, filteredMessages)
}

func apiMsgsPerUserHandler(c *gin.Context) {
	return
}

/*
GET and POST
if GET:

	return: all followers that :username follows

else if POST:

	if FOLLOW:
		make userA follow userB
		return: status code
	if UNFOLLOW:
		make userA unfollow userB
		return: status code
*/
func apiFllwsHandler(c *gin.Context) {
	return
}

// simulator logic is moved here now
// mathing the endpoints with the handlers for the simulator
// func registerSimulatorApi(router *gin.Engine) {

// 	router.GET("/", myTimelineHandler).Use(simulatorApi())
// 	router.GET("/public", publicTimelineHandler).Use(simulatorApi())
// 	router.GET("/:username", userTimelineHandler).Use(simulatorApi())
// 	router.GET("/register", registerHandler).Use(simulatorApi())
// 	router.GET("/login", loginHandler).Use(simulatorApi())
// 	router.GET("/logout", logoutHandler).Use(simulatorApi())
// 	router.GET("/:username/*action", userFollowActionHandler).Use(simulatorApi())

// 	router.POST("/register", registerHandler).Use(simulatorApi())
// 	router.POST("/login", loginHandler).Use(simulatorApi())
// 	router.POST("/add_message", addMessageHandler).Use(simulatorApi())

// }

// func simulatorApi() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		if c.GetHeader("simulatorApi") == "true" {
// 			c.Set("simulator", true)
// 		} else {
// 			c.Set("simulator", false)
// 		}
// 		c.Next()
// 	}
// }

// // use this function to make sure that the request is coming from the simulator
// func IsSimulatorRequest(c *gin.Context) bool {
// 	if isSumulator, exists := c.Get("simulator"); exists {
// 		return isSumulator.(bool)
// 	}
// 	return false
// }
