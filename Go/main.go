package main

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"html/template"
	"os"

	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

const (
	Database string = "./tmp/minitwit.db"
	Per_page int    = 30
)

func main() {

	// # configuration (0) _do we want them public or private?

	// var Debug bool = true
	// var Key string = "development key"

	//app aplication ?

	//using db connection (1)
	db, err := connect_db(Database)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//db initialization (2)
	// if err := init_db(db, "schema.sql"); err != nil {
	// 	log.Fatal(err)
	// }

	//after connecting to the db, we can define the handlers
	funcMap := template.FuncMap{
		"gravatar": gravatarURL,
	}

	// Load and parse your templates with the FuncMap
	tmpl := template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.tmpl"))

	// Create a Gin router and set the parsed templates
	router := gin.Default()
	router.SetHTMLTemplate(tmpl)
	router.Static("/static", "./static")

	// Define routes
	router.GET("/", timelineHandler)
	router.GET("/public", publicTimelineHandler)

	// Start the server
	router.Run(":8080")

}

func connect_db(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// db initialization (2)
// to keep separate functions of connection and initialization of the db. (here the db is structured with specific schema/format)
func init_db(db *sql.DB, schemaFile string) error {
	schema, err := os.ReadFile(schemaFile)
	if err != nil {
		return err
	}

	// Executing the schema SQL after it is being read in the previous step
	_, err = db.Exec(string(schema))
	if err != nil {
		return err
	}
	return err
}

// db query that returns list of dictionaries (3)
func query_db(db *sql.DB, query string, args []interface{}, one bool) ([]map[string]interface{}, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columns {
			columnPointers[i] = &values[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]interface{})
		for i, colName := range columns {
			val := columnPointers[i].(*interface{})
			rowMap[colName] = *val
		}

		result = append(result, rowMap)
		if one {
			break
		}
	}
	return result, nil
}

// // db check for username  and  returns the corresponding id (4)
// func get_user_id(db *sql.DB, username string) (error, *int) {
// 	query := "SELECT user_id FROM user WHERE username = ?"
// 	args := []interface{}{username}
// 	result, err := query_db(db, query, args, true)
// 	if err != nil {
// 		return err, nil
// 	}
// 	if len(result) == 0 {
// 		return nil, nil
// 	}
// 	return nil, result[0]["user_id"].(int)
// }

// endpoint handler (5)
func timelineHandler(c *gin.Context) {
	userID, exists := c.Get("userID") // You need to set userID in the context where you handle sessions
	if !exists {
		// Handle the case where the user is not logged in
		c.Redirect(http.StatusFound, "/public") // Redirect to public timeline
		return
	}

	query := `
    SELECT message.*, user.* FROM message, user
    WHERE message.flagged = 0 AND message.author_id = user.user_id AND (
        user.user_id = ? OR
        user.user_id IN (SELECT whom_id FROM follower WHERE who_id = ?))
    ORDER BY message.pub_date DESC LIMIT ?
    `
	var db, _ = connect_db(Database)
	// args := []interface{}{userID, userID, perPage}
	messages, err := query_db(db, query, nil, false)
	fmt.Println(messages)
	if err != nil {
		// Handle error
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// For template rendering with Gin
	c.HTML(http.StatusOK, "templates/timeline.html", gin.H{
		"messages": messages,
		"user":     userID, // Adjust according to how you manage users
	})
}

type User struct {
	UserID int

	// ... other user fields ...
}

type Message struct {
	MessageID int
	AuthorID  int
	Text      string
	PubDate   time.Time // Assuming it's a time.Time, format as needed
	// ... other message fields ...
	User         User
	Email        string
	Username     string
	Profile_link string
}

func publicTimelineHandler(c *gin.Context) {
	query := `
	SELECT message.*, user.* FROM message, user
	WHERE message.flagged = 0 AND message.author_id = user.user_id
	ORDER BY message.pub_date DESC LIMIT ?
	`
	var db, err = connect_db(Database)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer db.Close()

	args := []interface{}{Per_page}
	messages, err := query_db(db, query, args, false)
	// fmt.Println(messages)
	fmt.Println(err)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	var formattedMessages []Message
	for _, m := range messages {
		var msg Message
		// Use type assertion for int64, then convert to int
		if id, ok := m["message_id"].(int64); ok {
			msg.MessageID = int(id)
		}
		if authorID, ok := m["author_id"].(int64); ok {
			msg.AuthorID = int(authorID)
		}
		if userID, ok := m["user_id"].(int64); ok {
			msg.User.UserID = int(userID)
		}

		// For strings, direct type assertion is fine
		if text, ok := m["text"].(string); ok {
			msg.Text = text
		}
		if username, ok := m["username"].(string); ok {
			msg.Username = username
		}
		if email, ok := m["email"].(string); ok {
			msg.Email = email
		}
		link := "/" + msg.Username
		msg.Profile_link = strings.ReplaceAll(link, " ", "%20")
		formattedMessages = append(formattedMessages, msg)
	}

	c.HTML(http.StatusOK, "timeline.tmpl", gin.H{
		"Endpoint": "public_timeline", // or "user_timeline" etc, based on the context
		"Messages": formattedMessages,
	})

}

func gravatarURL(email string, size int) string {
	if size <= 0 {
		size = 80 // Default size
	}

	email = strings.ToLower(strings.TrimSpace(email))
	hash := md5.Sum([]byte(email))
	return fmt.Sprintf("http://www.gravatar.com/avatar/%x?d=identicon&s=%d", hash, size)
}
