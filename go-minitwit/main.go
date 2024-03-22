package main

import (
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
)

// todo: can we move these as well?
const (
	DATABASE string = "./tmp/minitwit_empty.db"
	PERPAGE  int    = 30
)

type FilteredMsg struct {
	Content string `json:"content"`
	PubDate int64  `json:"pub_date"`
	User    string `json:"user"`
}

var dbNew *gorm.DB

func main() {

	// Using db connection (1)
	var err error
	godotenv.Load()
	env := os.Getenv("EXECUTION_ENVIRONMENT")

	if env == "LOCAL" || env == "CI" {
		dbNew, err = connect_dev_DB("./tmp/minitwit_empty.db")
		if err != nil {
			panic("failed to connect to database")
		}

	} else {
		dbNew, err = connect_prod_DB()
		if err != nil {
			panic("failed to connect to database")
		}
	}

	// Create a Gin router and set the parsed templates
	router := gin.Default()
	router.Use(AfterRequest()) // This is the middleware that will be called after each request for Prometheus
	router.Use(beforeRequestHandler)
	router.Use(UserSignupMonitoring()) // This is the middleware that will be called after each request for Prometheus

	router.LoadHTMLGlob("./templates/*.html")

	// sessions, for cookies
	store := cookie.NewStore([]byte("devops"))
	router.Use(sessions.Sessions("session", store))

	// Static (styling)
	router.Static("/static", "./static")

	// Define routes -> Here is where the links are being registered! Check the html layout file
	// user routes
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

	// API routes
	// is it easier to separate the next two routes into two handlers?
	router.GET("/api/msgs", apiMsgsHandler)
	router.GET("/api/msgs/:username", apiMsgsPerUserHandler)
	router.GET("/api/fllws/:username", apiFllwsHandler)

	router.POST("/api/register", apiRegisterHandler)
	router.POST("/api/msgs/:username", apiMsgsPerUserHandler)
	router.POST("/api/fllws/:username", apiFllwsHandler)

	// some helper method to "cache" what was the latest simulator action
	router.GET("/api/latest", getLatest)

	// registering prometeus
	router.GET("/metrics", prometheusHandler())

	// Start the server
	router.Run(":8081")
}
