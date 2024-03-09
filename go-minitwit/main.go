package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/cpu"
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

// defining metrics -counter,cpu,responce time monitoring for Prometeus
var (
	cpuGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "minitwit_cpu_load_percent",
		Help: "Current load of the CPU in percent.",
	})
	responseCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "minitwit_http_responses_total",
		Help: "The count of HTTP responses sent.",
	})
	requestDurationSummary = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "minitwit_request_duration_milliseconds",
		Help: "Request duration distribution.",
	})
)

func main() {

	// Using db connection (1)
	var err error
	dbNew, err = connect_DB(DATABASE)
	if err != nil {
		panic("failed to connect to database")
	}

	// Create a Gin router and set the parsed templates
	router := gin.Default()
	router.Use(AfterRequest()) // This is the middleware that will be called after each request for Prometheus
	router.Use(beforeRequestHandler)

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

	//registerung prometeus
	router.GET("/metrics", prometheusHandler())

	// Start the server
	router.Run(":8081")

}

// defining registation of Prometeus
func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// for CPU use monitoring
func getCPUPercent() float64 {
	percentages, err := cpu.Percent(0, false)
	if err != nil {
		log.Printf("Error getting CPU usage: %v", err)
		return 0
	}
	if len(percentages) > 0 {
		return percentages[0]
	}
	return 0
}
func beforeRequestHandler(c *gin.Context) {
	// Set CPU usage
	cpuUsage := getCPUPercent()
	cpuGauge.Set(cpuUsage)
}

// for responce time monitoring and counting requests
func AfterRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This function needs to be deferred at the start of each handler where it's needed
		// Assuming db is the database connection passed through the context
		// And start time is set at the beginning of the handler

		defer func() {
			// Close the database connection
			if db, exists := c.Get("db"); exists {
				if dbConn, ok := db.(*sql.DB); ok {
					dbConn.Close()
				}
			}

			println("AfterRequest")

			// Increment the response counter for Prometeus
			responseCounter.Inc()

			// Calculate the elapsed time in milliseconds
			if startTime, exists := c.Get("startTime"); exists {
				if start, ok := startTime.(time.Time); ok {
					elapsedTime := time.Since(start).Milliseconds()
					requestDurationSummary.Observe(float64(elapsedTime))
				}
			}
		}()
	}
}

// Helper functions
// todo: move these to a "helperLibrary"
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
	hash := md5.Sum([]byte(userEnteredPwd))
	str := hex.EncodeToString(hash[:])
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

func formatMessages(messages []MessageUser) []MessageUI {
	var formattedMessages []MessageUI

	/*
		if reflect.TypeOf(m.Text).Kind() == reflect.String {
			filteredMsg.Content = m.Text
		}
	*/

	for _, m := range messages {
		var msg MessageUI
		// Use type assertion for int64, then convert to int
		if reflect.TypeOf(m.MessageID).Kind() == reflect.Int {
			msg.MessageID = m.MessageID
		}
		if reflect.TypeOf(m.AuthorID).Kind() == reflect.Int {
			msg.AuthorID = m.AuthorID
		}
		if reflect.TypeOf(m.UserID).Kind() == reflect.Int {
			msg.User.UserID = m.UserID
		}
		if reflect.TypeOf(m.Text).Kind() == reflect.String {
			msg.Text = m.Text
		}
		if reflect.TypeOf(m.Username).Kind() == reflect.String {
			msg.Username = m.Username
		}
		if reflect.TypeOf(m.Email).Kind() == reflect.String {
			msg.Email = m.Email
		}
		if reflect.TypeOf(m.PubDate).Kind() == reflect.Int {
			pubDateTime := time.Unix(int64(m.PubDate), 0)
			msg.PubDate = pubDateTime.Format("02/01/2006 15:04:05") // go time layout format is weird 1,2,3,4,5,6 ¬¬
		}
		link := "/" + msg.Username
		msg.Profile_link = strings.ReplaceAll(link, " ", "%20")

		gravatarURL := gravatarURL(msg.Email, 48)
		msg.Gravatar = gravatarURL

		formattedMessages = append(formattedMessages, msg)
	}

	return formattedMessages
}

func filterMessages(messages []MessageUser) []FilteredMsg {
	var filteredMessages []FilteredMsg
	for _, m := range messages {
		var filteredMsg FilteredMsg
		// content
		if reflect.TypeOf(m.Text).Kind() == reflect.String {
			filteredMsg.Content = m.Text
		}

		// publication date
		filteredMsg.PubDate = int64(m.PubDate)

		// user
		if reflect.TypeOf(m.Username).Kind() == reflect.String {
			filteredMsg.User = m.Username
		}

		filteredMessages = append(filteredMessages, filteredMsg)
	}
	return filteredMessages
}

func logMessage(message string) {
	// Specify the file path
	filePath := "./tmp/logging/logger.txt"

	// Open or create the file for writing
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	data := []byte(message)

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}
