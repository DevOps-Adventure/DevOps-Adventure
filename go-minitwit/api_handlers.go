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
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

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
	return
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
func apiMsgsHandler(c *gin.Context) {
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
