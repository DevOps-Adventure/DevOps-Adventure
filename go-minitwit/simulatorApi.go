// simulatorApi.go
package main

import (
	"github.com/gin-gonic/gin"
)

// mathing the endpoints with the handlers for the simulator
func registerSimulatorApi(router *gin.Engine) {

	//instead of using middleware, I am using separate handlers(specifically designed for simulator API) for cleaner code
	//	router.GET("/", myTimelineHandlerAPI).Use(simulatorApi())
	// Use API handlers specifically designed for the simulator
	router.GET("/", myTimelineHandlerAPI)
	router.GET("/public", publicTimelineHandlerAPI)
	router.GET("/:username", userTimelineHandlerAPI)
	router.GET("/register", registerHandlerAPI)
	router.GET("/login", loginHandlerAPI)
	router.GET("/logout", logoutHandlerAPI)
	router.POST("/login", loginHandlerAPI)
	router.POST("/add_message", addMessageHandlerAPI)
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
