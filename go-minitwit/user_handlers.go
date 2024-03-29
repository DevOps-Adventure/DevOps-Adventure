package main

import (
	"crypto/md5"
	"fmt"
	"strconv"

	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// Handlers
func userFollowActionHandler(c *gin.Context) {
	var err error
	session := sessions.Default(c)

	userID, errID := c.Cookie("UserID")
	if errID != nil {
		logger.WithFields(logrus.Fields{
			"source":   "user_interface",
			"endpoint": "following",
			"action":   "check login",
			"status":   "not logged in",
		}).Info("Attempt to follow/unfollow without being logged in")

		session.AddFlash("You need to login before continuing to follow or unfollow.")

		err = session.Save()
		if err != nil {
			panic("failed to save session in userFollowActionHandler")
		}

		c.Redirect(http.StatusFound, "/login")
		return

	}
	profileUserName := c.Param("username")
	profileUser, err := getUserByUsername(profileUserName)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"source":      "user_interface",
			"endpoint":    "following",
			"action":      "retrieve profile user",
			"status":      "failed",
			"profileUser": profileUserName,
			"error":       err.Error(),
		}).Error("Failed to retrieve user profile")

		fmt.Println("get user failed with:", err)
		c.Redirect(http.StatusFound, "/public")
		return
	}
	profileUserID := fmt.Sprintf("%v", profileUser.UserID)

	action := c.Param("action")

	if action == "/follow" {
		err = followUser(userID, profileUserID)
		if err != nil {
			panic("failed to followUser")
		}
    
		logger.WithFields(logrus.Fields{
			"source":   "user_interface",
			"endpoint": "following",
			"action":   "follow",
			"status":   "success",
		}).Info("User followed another user")
    
		session.AddFlash("You are now following " + profileUserName)
	}
	if action == "/unfollow" {
		err = unfollowUser(userID, profileUserID)
		if err != nil {
			panic("failed to unfollowUser")
		}

		logger.WithFields(logrus.Fields{
			"source":   "user_interface",
			"endpoint": "following",
			"action":   "follow",
			"status":   "success",
		}).Info("User followed another user")

		session.AddFlash("You are no longer following " + profileUserName)
	}
	err = session.Save()
	if err != nil {
		panic("failed to save session in userFollowActionHandler")
	}
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

			logger.WithFields(logrus.Fields{
				"source":   "user_interface",
				"endpoint": "public_timeline",
				"action":   "identify user",
			}).Info("User identified for public timeline")
		}
	}

	logger.WithFields(logrus.Fields{
		"source":        "user_interface",
		"endpoint":      "public_timeline",
		"action":        "render",
		"messagesCount": len(formattedMessages),
	}).Info("Rendering public timeline")

	// Render timeline template with the context including link variables
	c.HTML(http.StatusOK, "timeline.html", context)
}

func userTimelineHandler(c *gin.Context) {
	var err error
	session := sessions.Default(c)
	flashMessages := session.Flashes()
	err = session.Save()
	if err != nil {
		panic("failed to save session in userTimelineHandler")
	}
	profileUserName := c.Param("username")
	profileUser, err := getUserByUsername(profileUserName)

	if profileUser.Username == "" {

		logger.WithFields(logrus.Fields{
			"source":   "user_interface",
			"endpoint": "user_timeline",
			"action":   "fetch_user",
			"status":   "user_not_found",
		}).Warn("User not found for timeline")

		c.AbortWithStatus(404)
		return
	}
	if err != nil {

		logger.WithFields(logrus.Fields{
			"source":   "user_interface",
			"endpoint": "user_timeline",
			"action":   "fetch_user",
			"status":   "error",
			"error":    err.Error(),
		}).Error("Error fetching user for timeline")

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

			logger.WithFields(logrus.Fields{
				"source":   "user_interface",
				"endpoint": "user_timeline",
				"action":   "check_follow_status",
				"status":   "error",
				"error":    err.Error(),
			}).Error("Error checking follow status")

			logMessage(err.Error())
			return
		}
	}

	messages, err := getUserMessages(pUserId, PERPAGE)
	fmt.Println(messages)

	if err != nil {

		logger.WithFields(logrus.Fields{
			"source":   "user_interface",
			"endpoint": "user_timeline",
			"action":   "fetch_user_messages",
			"status":   "error",
			"error":    err.Error(),
		}).Error("Error fetching user messages")

    //nolint:all
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	formattedMessages := formatMessages(messages)

	logger.WithFields(logrus.Fields{
		"source":         "user_interface",
		"endpoint":       "user_timeline",
		"action":         "render_user_timeline",
		"messages_count": len(formattedMessages),
	}).Info("Rendering users public timeline")

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
	var err error
	userID, err := c.Cookie("UserID")
	errMsg := c.Query("error")

	if err != nil {

		logger.WithFields(logrus.Fields{
			"source":   "user_interface",
			"endpoint": "my_timeline",
			"action":   "get_cookie",
			"status":   "error",
			"error":    err.Error(),
		}).Error("Error getting user information")

		c.Redirect(http.StatusFound, "/public")
		return
	}

	userName, err := getUserNameByUserID(userID)

	if err != nil {
		logger.WithFields(logrus.Fields{
			"source":   "user_interface",
			"endpoint": "my_timeline",
			"action":   "get_user_name",
			"status":   "error",
			"error":    err.Error(),
		}).Error("Error getting username by id")

    //nolint:all
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	session := sessions.Default(c)
	flashMessages := session.Flashes()
	err = session.Save() // Clear flashes after retrieving
	if err != nil {
		panic("failed to save session in myTimelineHandler")
	}

	messages, err := getMyMessages(userID)

	if err != nil {
		logger.WithFields(logrus.Fields{
			"source":   "user_interface",
			"endpoint": "my_timeline",
			"action":   "get_my_messages",
			"status":   "error",
			"error":    err.Error(),
		}).Error("Error getting users messages")

		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	formattedMessages := formatMessages(messages)
	fmt.Println(formattedMessages)

	logger.WithFields(logrus.Fields{
		"source":         "user_interface",
		"endpoint":       "my_timeline",
		"action":         "format_messages",
		"messages_count": len(formattedMessages),
	}).Info("Rendering users timeline")

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
	var err error

	session := sessions.Default(c)

	userID, err := c.Cookie("UserID")
	userIDString, errStr := strconv.Atoi(userID)
	if err != nil || errStr != nil {

		logger.WithFields(logrus.Fields{
			"source":   "user_interface",
			"endpoint": "add_messages",
			"action":   "get_cookie",
			"status":   "error",
			"error":    err.Error(),
		}).Error("Error getting user information")

		c.Redirect(http.StatusFound, "/public")
		return
	}

	var errorData string
	if c.Request.Method == http.MethodPost {
		err := c.Request.ParseForm()
		if err != nil {

			logger.WithFields(logrus.Fields{
				"source":   "user_interface",
				"endpoint": "add_messages",
				"action":   "parse_from_data",
				"status":   "error",
				"error":    err.Error(),
			}).Error("Error failed to parse form data")

			errorData = "Failed to parse form data"
			c.Redirect(http.StatusBadRequest, "/?error="+errorData)
			return
		}

		// Validate form data
		text := c.Request.FormValue("text")

		if text == "" {
			c.Redirect(http.StatusSeeOther, "/")
			session.AddFlash("You have to enter a value")
			err = session.Save()
			if err != nil {
				panic("failed to save session in addMessageHandler")
			}
			return
		} else {
			err := addMessage(text, userIDString)
			if err != nil {

				logger.WithFields(logrus.Fields{
					"source":   "user_interface",
					"endpoint": "add_messages",
					"action":   "enter_value",
					"status":   "error",
					"error":    err.Error(),
				}).Error("Error failed to add message")

				errorData = "Failed to add message"
				c.Redirect(http.StatusInternalServerError, "/?error="+errorData)
				return
			}

			logger.WithFields(logrus.Fields{
				"source":   "user_interface",
				"endpoint": "add_messages",
				"action":   "enter_value",
				"status":   "success",
			}).Info("Rendering users timeline")

			c.Redirect(http.StatusSeeOther, "/")
			session.AddFlash("Your message was recorded")
			err = session.Save()
			if err != nil {
				panic("failed to save session in addMessageHandler")
			}
			return
		}
	}
	c.Redirect(http.StatusSeeOther, "/")
}

func registerHandler(c *gin.Context) {

	session := sessions.Default(c)

	userID, exists := c.Get("UserID")
	if exists {

		logger.WithFields(logrus.Fields{
			"source":   "user_interface",
			"endpoint": "register_user",
			"action":   "get_user",
		}).Info("User exists")

		fmt.Println("userID:", userID)
		return
	}

	var errorData string
	if c.Request.Method == http.MethodPost {
		err := c.Request.ParseForm()
		if err != nil {

			logger.WithFields(logrus.Fields{
				"source":   "user_interface",
				"endpoint": "register_user",
				"action":   "parse_data",
				"status":   "error",
				"error":    err.Error(),
			}).Error("Error failed to parse form data")

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
			logger.WithFields(logrus.Fields{
				"source":   "user_interface",
				"endpoint": "register_user",
				"action":   "get_user_by_id",
				"status":   "error",
				"error":    err.Error(),
			}).Error("Error getting username by id")
      
      //nolint:all
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

				logger.WithFields(logrus.Fields{
					"source":   "user_interface",
					"endpoint": "register_user",
					"action":   "registration_attempt",
					"status":   "failed",
					"reason":   "error_registering_user",
					"error":    err.Error(),
				}).Error("Failed registration attempt due to an error during registration")

				errorData = "Failed to register user"
				c.HTML(http.StatusInternalServerError, "register.html", gin.H{
					"RegisterBody": true,
					"Error":        errorData,
				})
				return
			}

			logger.WithFields(logrus.Fields{
				"source":   "user_interface",
				"endpoint": "register_user",
				"action":   "registration",
				"status":   "success",
			}).Info("User successfully registered")

			// Redirect to login page after successful registration
			session.AddFlash("You were successfully registered and can login now")
			// print session info
			fmt.Println("session info:", session, "Logged in")
			err = session.Save()
			if err != nil {
				panic("failed to save session in registerHandler")
			}
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
	var err error
	session := sessions.Default(c)
	flashMessages := session.Flashes()
	err = session.Save()
	if err != nil {
		panic("failed to save session in loginHandler")
	}

	userID, _ := c.Cookie("UserID")
	if userID != "" {

		logger.WithFields(logrus.Fields{
			"source":   "user_interface",
			"endpoint": "login_user",
			"action":   "login_check",
			"status":   "already_logged_in",
			"userID":   userID,
		}).Info("User already logged in, redirecting")

		session.AddFlash("You were logged in")
		err = session.Save()
		if err != nil {
			panic("failed to save session in loginHandler")
		}
		c.Redirect(http.StatusFound, "/")
		return
	}

	var errorData string

	if c.Request.Method == http.MethodPost {

		err := c.Request.ParseForm()
		if err != nil {

			logger.WithFields(logrus.Fields{
				"source":   "user_interface",
				"endpoint": "login_user",
				"action":   "login_attempt",
				"status":   "failed",
				"reason":   "parse_form_error",
				"error":    err.Error(),
			}).Error("Failed to parse login form")

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
			//nolint:all
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
				logger.WithFields(logrus.Fields{
					"source":   "user_interface",
					"endpoint": "login_user",
					"action":   "login_attempt",
					"status":   "failed",
					"reason":   "user_id_retrieval_error",
					"error":    err.Error(),
				}).Error("Failed to retrieve userID during login")
        
        //nolint:all
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			logger.WithFields(logrus.Fields{
				"source":   "user_interface",
				"endpoint": "login_user",
				"action":   "login",
				"status":   "success",
			}).Info("User successfully logged in")

			c.SetCookie("UserID", fmt.Sprint(userID), 3600, "/", "", false, true)
			session.AddFlash("You were logged in")
			err = session.Save()
			if err != nil {
				panic("failed to save session in loginHandler")
			}
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
	var err error
	session := sessions.Default(c)

	logger.WithFields(logrus.Fields{
		"source":   "user_interface",
		"endpoint": "logout_user",
		"action":   "logout",
		"status":   "success",
	}).Info("User successfully logged out")

	session.AddFlash("You were logged out")
	err = session.Save()
	if err != nil {
		panic("failed to save session in logoutHandler")
	}
	// Invalidate the cookie by setting its max age to -1
	// will delete the cookie <- nice stuff
	c.SetCookie("UserID", "", -1, "/", "", false, true)
	// redirect the user to the home page or login page
	c.Redirect(http.StatusFound, "/login")
}
