// simulatorApi.go
package main

import (
	"github.com/gin-gonic/gin"
)

// mathing the endpoints with the handlers for the simulator
func registerSimulatorApi(router *gin.Engine) {

	router.GET("/", myTimelineHandler).Use(simulatorApi())
	router.GET("/public", publicTimelineHandler).Use(simulatorApi())
	router.GET("/:username", userTimelineHandler).Use(simulatorApi())
	router.GET("/register", registerHandler).Use(simulatorApi())
	router.GET("/login", loginHandler).Use(simulatorApi())
	router.GET("/logout", logoutHandler).Use(simulatorApi())
	router.GET("/:username/*action", userFollowActionHandler).Use(simulatorApi())

	router.POST("/register", registerHandler).Use(simulatorApi())
	router.POST("/login", loginHandler).Use(simulatorApi())
	router.POST("/add_message", addMessageHandler).Use(simulatorApi())

}

func simulatorApi() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("simulatorApi") == "true" {
			c.Set("simulator", true)
		} else {
			c.Set("simulator", false)
		}
		c.Next()
	}
}

// use this function to make sure that the request is coming from the simulator
func IsSimulatorRequest(c *gin.Context) bool {
	if isSumulator, exists := c.Get("simulator"); exists {
		return isSumulator.(bool)
	}
	return false
}
