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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// struct for error data (need to be JSON before return)
type ErrorData struct {
	status    int
	error_msg string
}

type UserData struct {
	Username string
	Email    string
	Pwd      string
}

type MessageData struct {
	Content string `json:"content"`
}

func not_req_from_simulator(c *gin.Context) (statusCode int, errStr string) {
	auth := c.Request.Header.Get("Authorization")
	if auth != "Basic c2ltdWxhdG9yOnN1cGVyX3NhZmUh" {
		statusCode = 403
		errStr = "You are not authorized to use this resource!"
		return statusCode, errStr
	}
	return
}

func updateLatestHandler(c *gin.Context) {
	parsedCommandID := c.Query("latest")
	commandID, err := strconv.Atoi(parsedCommandID)

	if err != nil || parsedCommandID == "" {
		// Handle the case where the parameter is not present or cannot be converted to an integer
		commandID = -1
	}
	if commandID != -1 {
		err := updateLatest(commandID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update latest value"})
			return
		}
	}
}

func getLatestHandler(c *gin.Context) {
	latestProcessedCommandID, err := getLatest()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read latest value"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"latest": latestProcessedCommandID})
}

func getLatestHelper() int {
	latestProcessedCommandID, err := getLatest()
	if err != nil {
		return -2
	}
	return latestProcessedCommandID
}

/*
/api/register
POST
Takes data from the POST and registers a user in the db
returns: ("", 204) or ({"status": 400, "error_msg": error}, 400)
*/
func apiRegisterHandler(c *gin.Context) {

	updateLatestHandler(c)
	latest := getLatestHelper()
	logMessage(fmt.Sprint(latest) + " apiRegisterHandler: registering user.")

	errorData := ErrorData{
		status:    0,
		error_msg: "",
	}

	//Check if user already exists
	userID, exists := c.Get("UserID")
	if exists {

		logger.WithFields(logrus.Fields{
			"source":   "api",
			"endpoint": "/api/register",
			"action":   "check_user_exists",
		}).Warn("Attempt to register an existing user")

		errorData.status = 400
		errorData.error_msg = "User already exists: " + fmt.Sprintf("%v", userID)
		c.AbortWithStatusJSON(400, errorData)
		return
	}

	if c.Request.Method == http.MethodPost {
		// Read the request body
		var registerReq UserData
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {

			logger.WithFields(logrus.Fields{
				"source":   "api",
				"endpoint": "/api/register",
				"action":   "read_request_body",
				"status":   "error",
				"error":    err.Error(),
			}).Error("Error failed to read request body")

			errorData.status = 400
			errorData.error_msg = "Failed to read JSON"
			c.AbortWithStatusJSON(400, errorData)
			return
		}

		// Parse the request body from JSON
		// Unmarshal parses the JSON and stores it in a pointer (registerReq)
		if err := json.Unmarshal(body, &registerReq); err != nil {

			logger.WithFields(logrus.Fields{
				"source":   "api",
				"endpoint": "/api/register",
				"action":   "parse_json",
				"status":   "error",
				"error":    err.Error(),
			}).Error("Failed to parse request JSON")

			errorData.status = 400
			errorData.error_msg = "Failed to parse JSON"
			c.AbortWithStatusJSON(400, errorData)
			return
		}

		//Set the user data
		username := registerReq.Username
		email := registerReq.Email
		password := registerReq.Pwd

		// Get user ID
		userID, err := getUserIDByUsername(username)
		if err != nil {

			logger.WithFields(logrus.Fields{
				"source":   "api",
				"endpoint": "/api/register",
				"action":   "get_user_by_id",
				"status":   "error",
				"error":    err.Error(),
			}).Error("Error getting username by id")

			errorData.status = 400
			errorData.error_msg = "Failed to get userID"
			c.AbortWithStatusJSON(400, errorData)
			return
		}

		// Check for errors
		if username == "" {
			errorData.status = 400
			errorData.error_msg = "You have to enter a username"
			c.AbortWithStatusJSON(400, errorData.error_msg)
			return

		} else if email == "" || !strings.Contains(email, "@") {
			errorData.status = 400
			errorData.error_msg = "You have to enter a valid email address"
			c.AbortWithStatusJSON(400, errorData.error_msg)
			return

		} else if password == "" {
			errorData.status = 400
			errorData.error_msg = "You have to enter a password"
			c.AbortWithStatusJSON(400, errorData.error_msg)
			return

		} else if fmt.Sprint(userID) != "-1" {
			errorData.status = 400
			errorData.error_msg = "The username is already taken"
			c.AbortWithStatusJSON(400, errorData.error_msg)
			return

		} else {
			hash := md5.Sum([]byte(password))
			err := registerUser(username, email, hash)
			if err != nil {

				logger.WithFields(logrus.Fields{
					"source":   "api",
					"endpoint": "/api/register",
					"action":   "registration_attempt",
					"status":   "failed",
					"reason":   "error_registering_user",
					"error":    err.Error(),
				}).Error("Failed registration attempt due to an error during registration")

				errorData.status = 400
				errorData.error_msg = "Failed to register user"
				c.AbortWithStatusJSON(400, errorData.error_msg)
				return
			}
		}

		if errorData.error_msg != "" {
			c.AbortWithStatusJSON(400, errorData.error_msg)
			return
		} else {

			logger.WithFields(logrus.Fields{
				"source":   "api",
				"endpoint": "/api/register",
				"action":   "registration",
				"status":   "success",
			}).Info("User successfully registered")

			c.JSON(204, "")
		}
	}
}

/*
/api/msgs
/api/msgs?no=<num>
*/
func apiMsgsHandler(c *gin.Context) {

	updateLatestHandler(c)
	latest := getLatestHelper()
	logMessage(fmt.Sprint(latest) + " apiMsgsHandler: getting all messages.")

	errorData := ErrorData{
		status:    0,
		error_msg: "",
	}

	not_req_from_sim_statusCode, not_req_from_sim_errStr := not_req_from_simulator(c)
	if not_req_from_sim_statusCode == 403 && not_req_from_sim_errStr != "" {

		logger.WithFields(logrus.Fields{
			"source":   "api",
			"endpoint": "/api/messages",
			"action":   "access_denied",
			"reason":   not_req_from_sim_errStr,
		}).Warn("Request denied: not from simulator")

		errorData.status = http.StatusForbidden
		errorData.error_msg = not_req_from_sim_errStr
		c.AbortWithStatusJSON(http.StatusForbidden, errorData.error_msg)
		return
	}

	numMsgs := c.Request.Header.Get("no")
	numMsgsInt, err := strconv.Atoi(numMsgs)
	// fallback on default value
	if err != nil {

		logger.WithFields(logrus.Fields{
			"source":   "api",
			"endpoint": "/api/messages",
			"action":   "parse_header_fallback",
			"numMsgs":  numMsgs,
			"fallback": 100,
			"error":    err.Error(),
		}).Info("Falling back to default number of messages due to parsing error")

		numMsgsInt = 100
	}

	messages, err := getPublicMessages(numMsgsInt)
	if err != nil {

		logger.WithFields(logrus.Fields{
			"source":   "api",
			"endpoint": "/api/messages",
			"action":   "fetch_messages",
			"status":   "error",
			"error":    err.Error(),
		}).Error("Failed to fetch messages from DB")

		errorData.status = http.StatusBadRequest
		errorData.error_msg = "Failed to fetch messages from DB"
		c.AbortWithStatusJSON(http.StatusBadRequest, errorData)
	}

	filteredMessages := filterMessages(messages)
	jsonFilteredMessages, _ := json.Marshal(filteredMessages)
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, string(jsonFilteredMessages))
}

/*
/api/msgs/<username>
*/
func apiMsgsPerUserHandler(c *gin.Context) {

	updateLatestHandler(c)
	latest := getLatestHelper()
	logMessage(fmt.Sprint(latest) + " apiMsgsPerUserHandler: getting all messages by user " + c.Param("username") + ".")

	errorData := ErrorData{
		status:    0,
		error_msg: "",
	}

	not_req_from_sim_statusCode, not_req_from_sim_errStr := not_req_from_simulator(c)
	if not_req_from_sim_statusCode == 403 && not_req_from_sim_errStr != "" {

		logger.WithFields(logrus.Fields{
			"source":   "api",
			"endpoint": "/api/messages_per_user",
			"action":   "access_denied",
			"reason":   not_req_from_sim_errStr,
		}).Warn("Request denied: not from simulator")

		errorData.status = http.StatusForbidden
		errorData.error_msg = not_req_from_sim_errStr
		c.AbortWithStatusJSON(http.StatusForbidden, errorData.error_msg)
		return
	}

	profileUserName := c.Param("username")
	userId, err := getUserIDByUsername(profileUserName)
	if userId == -1 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if c.Request.Method == http.MethodGet {
		numMsgs := c.Request.Header.Get("no")
		numMsgsInt, err := strconv.Atoi(numMsgs)
		// fallback on default value
		if err != nil {
			numMsgsInt = 100

			logger.WithFields(logrus.Fields{
				"source":   "api",
				"endpoint": "/api/messages_per_user",
				"action":   "parse_header_fallback",
				"fallback": numMsgsInt,
			}).Info("Fallback to default number of messages due to parsing error")
		}

		messages, err := getUserMessages(userId, numMsgsInt)
		if err != nil {

			logger.WithFields(logrus.Fields{
				"source":   "api",
				"endpoint": "/api/messages_per_user",
				"action":   "fetch_messages",
				"status":   "error",
				"error":    err.Error(),
			}).Error("Failed to fetch messages from DB")

			errorData.status = http.StatusBadRequest
			errorData.error_msg = "Failed to fetch messages from DB"
			c.AbortWithStatusJSON(http.StatusInternalServerError, errorData)
		}

		// Log successful retrieval of messages
		logger.WithFields(logrus.Fields{
			"source":     "api",
			"endpoint":   "/api/messages_per_user",
			"action":     "retrieve_messages",
			"status":     "success",
			"numMsgs":    numMsgsInt,
			"numResults": len(messages),
		}).Info("Successfully retrieved messages")

		filteredMessages := filterMessages(messages)
		jsonFilteredMessages, _ := json.Marshal(filteredMessages)
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, string(jsonFilteredMessages))

	} else if c.Request.Method == http.MethodPost {
		// Read the request body
		var messageReq MessageData
		body, err := io.ReadAll(c.Request.Body)

		if err != nil {

			logger.WithFields(logrus.Fields{
				"source":   "api",
				"endpoint": "/api/messages_per_user",
				"action":   "read_request_body",
				"status":   "error",
				"error":    err.Error(),
			}).Error("Failed to read request body")

			errorData.status = 400
			errorData.error_msg = "Failed to read JSON"
			c.AbortWithStatusJSON(http.StatusBadRequest, errorData)
		}

		if err := json.Unmarshal(body, &messageReq); err != nil {

			logger.WithFields(logrus.Fields{
				"source":   "api",
				"endpoint": "/api/messages_per_user",
				"action":   "parse_json",
				"status":   "error",
				"error":    err.Error(),
			}).Error("Failed to parse JSON body")

			errorData.status = 400
			errorData.error_msg = "Failed to parse JSON"
		}

		text := messageReq.Content
		fmt.Println(text)
		authorId, err := getUserIDByUsername(profileUserName)
		if err != nil {
			errorData.status = http.StatusBadRequest
			errorData.error_msg = "Failed to get userID"
			c.AbortWithStatusJSON(http.StatusBadRequest, errorData)
		}

		err = addMessage(text, authorId)
		if err != nil {
			errorData.status = http.StatusInternalServerError
			errorData.error_msg = "Failed to upload message"
			c.AbortWithStatusJSON(http.StatusInternalServerError, errorData)
		}

		logger.WithFields(logrus.Fields{
			"source":   "api",
			"endpoint": "/api/messages_per_user",
			"action":   "upload_message",
			"status":   "success",
		}).Info("Successfully uploaded message")

		c.String(http.StatusNoContent, "")
	}
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

/api/fllws/<username>
*/
func apiFllwsHandler(c *gin.Context) {

	updateLatestHandler(c)
	latest := getLatestHelper()
	logMessage(fmt.Sprint(latest) + " apiFllwsHandler: checking follow")

	errorData := ErrorData{
		status:    0,
		error_msg: "",
	}

	not_req_from_sim_statusCode, not_req_from_sim_errStr := not_req_from_simulator(c)
	if not_req_from_sim_statusCode == 403 && not_req_from_sim_errStr != "" {

		logger.WithFields(logrus.Fields{
			"source":   "api",
			"endpoint": "/api/fllw",
			"action":   "access_denied",
			"reason":   not_req_from_sim_errStr,
		}).Warn("Request denied: not from simulator")

		errorData.status = http.StatusForbidden
		errorData.error_msg = not_req_from_sim_errStr
		c.AbortWithStatusJSON(http.StatusForbidden, errorData.error_msg)
		return
	}

	if c.Request.Method == http.MethodGet {
		profileUserName := c.Param("username")
		numFollr := c.Request.Header.Get("no")
		numFollrInt, err := strconv.Atoi(numFollr)
		// fallback on default value
		if err != nil {
			logger.WithFields(logrus.Fields{
				"source":   "api",
				"endpoint": "/api/fllw",
				"action":   "parse_header_fallback",
				"fallback": numFollrInt,
			}).Info("Fallback to default number of followers due to parsing error")
			numFollrInt = 100
		}

		userId, err := getUserIDByUsername(profileUserName)
		if err != nil || userId == -1 {

			logger.WithFields(logrus.Fields{
				"source":   "api",
				"endpoint": "/api/fllw",
				"action":   "get_user_id",
				"status":   "failed",
			}).Error("Failed to get user ID for follow/unfollow actions")

			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		// Fetch all followers for the user
		userIdStr := strconv.Itoa(userId)
		followers, err := getFollowing(userIdStr, numFollrInt)
		if err != nil {

			logger.WithFields(logrus.Fields{
				"source":   "api",
				"endpoint": "/api/fllw",
				"action":   "fetch_followers",
				"status":   "error",
				"error":    err.Error(),
			}).Error("Failed to fetch followers from DB")

			errorData.status = http.StatusInternalServerError
			errorData.error_msg = "Failed to fetch followers from DB"
			c.AbortWithStatusJSON(http.StatusInternalServerError, errorData)
		}

		// Successfully retrieved followers, log this event
		logger.WithFields(logrus.Fields{
			"source":     "api",
			"endpoint":   "/api/fllw",
			"action":     "retrieve_followers",
			"status":     "success",
			"numResults": len(followers),
		}).Info("Successfully retrieved followers")

		// empty slice for follower usernames
		followerNames := []string{}

		// Append the usernames to the followerNames slice
		for _, follower := range followers {
			followerNames = append(followerNames, string(follower.Username))
		}

		// Prepare response
		followersResponse := gin.H{
			"follows": followerNames,
		}

		// Send JSON response of all followers
		c.JSON(200, followersResponse)

	} else if c.Request.Method == http.MethodPost {
		// POST request
		var requestBody struct {
			Follow   string `json:"follow"`
			Unfollow string `json:"unfollow"`
		}

		// Bind JSON data to requestBody
		if err := c.BindJSON(&requestBody); err != nil {

			logger.WithFields(logrus.Fields{
				"source":   "api",
				"endpoint": "/api/fllw",
				"action":   "bind_json",
				"status":   "error",
				"error":    err.Error(),
			}).Error("Failed to bind JSON for follow/unfollow action")

			errorData.status = http.StatusNotFound
			errorData.error_msg = "Failed to parse JSON"
			c.AbortWithStatusJSON(http.StatusNotFound, errorData)
			return
		}

		profileUserName := c.Param("username")

		// Convert profileUserName to userID
		userId, err := getUserIDByUsername(profileUserName)
		if err != nil || userId == -1 {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		userIdStr := strconv.Itoa(userId)

		if requestBody.Follow != "" {
			// Follow logic
			// Convert requestBody.Follow to profileUserID
			profileUserID, err := getUserIDByUsername(requestBody.Follow)
			if err != nil || profileUserID == -1 {
				logger.WithFields(logrus.Fields{
					"source":   "api",
					"endpoint": "/api/fllw",
					"action":   "get_user_id",
					"status":   "failed",
				}).Error("Failed to get user ID for follow/unfollow actions")
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			profileUserIDStr := strconv.Itoa(profileUserID)

			// Follow the user
			if err := followUser(userIdStr, profileUserIDStr); err != nil {

				logger.WithFields(logrus.Fields{
					"source":   "api",
					"endpoint": "/api/fllw",
					"action":   "follow_user",
					"status":   "failed",
				}).Error("Failed to follow user")

				errorData.status = http.StatusNotFound
				errorData.error_msg = "Failed to follow user"
				c.AbortWithStatusJSON(http.StatusNotFound, errorData)
				return
			}

			logger.WithFields(logrus.Fields{
				"source":   "api",
				"endpoint": "/api/fllw",
				"action":   "follow",
				"status":   "requested",
				"follower": profileUserName,
				"followee": requestBody.Follow,
			}).Info("Follow request processed")

			c.JSON(http.StatusNoContent, "")
			return
		} else if requestBody.Unfollow != "" {
			// Unfollow logic
			// Convert requestBody.Unfollow to profileUserID
			profileUserID, err := getUserIDByUsername(requestBody.Unfollow)
			if err != nil || profileUserID == -1 {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			profileUserIDStr := strconv.Itoa(profileUserID)

			// Unfollow the user
			if err := unfollowUser(userIdStr, profileUserIDStr); err != nil {
				logger.WithFields(logrus.Fields{
					"source":   "api",
					"endpoint": "/api/fllw",
					"action":   "unfollow_user",
					"status":   "failed",
				}).Error("Failed to unfollow user")

				errorData.status = http.StatusNotFound
				errorData.error_msg = "Failed to unfollow user"
				c.AbortWithStatusJSON(http.StatusNotFound, errorData)
				return
			}

			logger.WithFields(logrus.Fields{
				"source":   "api",
				"endpoint": "/api/fllw",
				"action":   "unfollow",
				"status":   "requested",
				"follower": profileUserName,
				"followee": requestBody.Unfollow,
			}).Info("Unfollow request processed")

			c.JSON(http.StatusNoContent, "")
		} else {
			errorData.status = http.StatusNotFound
			errorData.error_msg = "No 'follow' or 'unfollow' provided in request"
			c.AbortWithStatusJSON(http.StatusNotFound, errorData)
			return
		}
	}
}
