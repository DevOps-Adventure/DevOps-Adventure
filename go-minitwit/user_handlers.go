package main

import (
	"crypto/md5"
	"fmt"
	"strconv"

	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

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
	profileUserID := fmt.Sprintf("%v", profileUser.UserID)

	action := c.Param("action")

	if action == "/follow" {
		followUser(userID, profileUserID)
		session.AddFlash("You are now following " + profileUserName)
	}
	if action == "/unfollow" {
		unfollowUser(userID, profileUserID)
		session.AddFlash("You are no longer following " + profileUserName)
	}
	session.Save()
	c.Redirect(http.StatusFound, "/"+profileUserName)
}

func publicTimelineHandler(c *gin.Context) {

	// need to pass a default value to getPublicMessages (GoLang doesn't support default values for arguments)
	messages, err := getPublicMessages(PERPAGE)
	if err != nil {
		return
	}
	formattedMessages := formatMessages(messages)

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

	if profileUser.Username == "" {
		c.AbortWithStatus(404)
		return
	}
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// does the logged in user follow them
	followed := false
	pUserId := profileUser.UserID
	profileName := profileUser.Username
	userID, errID := c.Cookie("UserID")
	userIDInt, _ := strconv.Atoi(userID)
	userName, _ := getUserNameByUserID(userID)

	if errID == nil {
		followed, err = checkFollowStatus(userIDInt, pUserId)
		if err != nil {
			logMessage(err.Error())
			return
		}
	}

	messages, err := getUserMessages(pUserId, PERPAGE)
	fmt.Println(messages)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	formattedMessages := formatMessages(messages)

	c.HTML(http.StatusOK, "timeline.html", gin.H{
		"TimelineBody":    true,
		"Endpoint":        "user_timeline",
		"UserID":          userIDInt,
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
	errMsg := c.Query("error")

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

	formattedMessages := formatMessages(messages)
	fmt.Println(formattedMessages)

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
		"Error":        errMsg,
	})
}

func addMessageHandler(c *gin.Context) {

	session := sessions.Default(c)

	userID, err := c.Cookie("UserID")
	userIDString, errStr := strconv.Atoi(userID)
	if err != nil || errStr != nil {
		c.Redirect(http.StatusFound, "/public")
		return
	}

	var errorData string
	if c.Request.Method == http.MethodPost {
		err := c.Request.ParseForm()
		if err != nil {
			errorData = "Failed to parse form data"
			c.Redirect(http.StatusBadRequest, "/?error="+errorData)
			return
		}

		// Validate form data
		text := c.Request.FormValue("text")

		if text == "" {
			c.Redirect(http.StatusSeeOther, "/")
			session.AddFlash("You have to enter a value")
			session.Save()
			return
		} else {
			err := addMessage(text, userIDString)
			if err != nil {
				errorData = "Failed to add message"
				c.Redirect(http.StatusInternalServerError, "/?error="+errorData)
				return
			}

			c.Redirect(http.StatusSeeOther, "/")
			session.AddFlash("Your message was recorded")
			session.Save()
			return
		}
	}
	c.Redirect(http.StatusSeeOther, "/")
}

func registerHandler(c *gin.Context) {

	session := sessions.Default(c)

	userID, exists := c.Get("UserID")
	if exists {
		fmt.Println("userID:", userID)
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
		} else if fmt.Sprint(userID) != "-1" {
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
			// print session info
			fmt.Println("session info:", session, "Logged in")
			session.Save()
			c.Redirect(http.StatusSeeOther, "/login")
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

		if user.Username == "" {
			errorData = "Invalid username"
		} else if !checkPasswordHash(password, user.PwHash) {
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
