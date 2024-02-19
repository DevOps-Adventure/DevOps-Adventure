// db_connection.go
// 2 in function is used to differentiate from the main.go file - needs to be removed when final
package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3" // Import the SQLite3 driver
)

// Global constants (if needed, else define them where they are used)
const (
	DATABASE2 string = "./tmp/minitwit.db"
)

// connect_db creates and returns a new database connection
func connect_db2(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// to keep separate functions of connection and initialization of the db. (here the db is structured with specific schema/format)
func init_db2(db *sql.DB, schemaFile string) error {
	schema, err := os.ReadFile(schemaFile)
	if err != nil {
		return err
	}
	_, err = db.Exec(string(schema))
	return err
}

// query_db executes a query on the database and returns a list of dictionaries (maps)
// db query that returns list of dictionaries (3)
func query_db2(db *sql.DB, query string, args []interface{}, one bool) ([]map[string]interface{}, error) {
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

// addUser adds a new user to the database
// needs to be implemented
// example usage of query_db
func addUser(db *sql.DB, username, email, hashedPassword string) error {
	// query to insert a new user
	query := `INSERT INTO user (username, email, pw_hash) VALUES (?, ?, ?)`

	// statement with the user details
	_, err := db.Exec(query, username, email, hashedPassword)
	if err != nil {
		return fmt.Errorf("addUser: %v", err)
	}

	return nil
}

// getUserByID fetches a user by their ID
// (Copy from existing code)
func getUserIDByUsername2(userName string) (int64, error) {
	return 0, nil
}

// checkUserCredentials verifies a user login details
// (Copy from existing code)
func getUserNameByUserID2(userID string) (string, error) {
	return "", nil
}

// addMessage adds a new message to the database
// (Copy from existing code)
func addMessage2(text string, author_id string) error {
	query := `insert into message (author_id, text, pub_date, flagged) values (?, ?, ?, 0)`
	var db, err = connect_db(DATABASE)
	if err != nil {
		fmt.Println("error in addMessage query")
		return err
	}
	currentTime := time.Now().UTC()
	unixTimestamp := currentTime.Unix()

	args := []interface{}{author_id, text, unixTimestamp, 0}
	query_db2(db, query, args, false)
	return err
}

// getPublicMessages fetches messages for display.
// (Copy from existing code)
func getPublicMessages(db *sql.DB, limit int) ([]Message, error) {
	return nil, nil
}

// registerUser registers a new user
// (Copy from existing code)
func registerUser2(userName string, email string, password [16]byte) error {
	query := `insert into user (username, email, pw_hash) values (?, ?, ?)`
	var db, err = connect_db(DATABASE)
	if err != nil {
		return err
	}
	args := []interface{}{userName, email, pq.Array(password)}
	messages, err := query_db2(db, query, args, false)
	fmt.Println("this is the messages", messages)
	return err
}

// followUser adds a new follower to the database
// (Copy from existing code)
func followUser2(userID string, profileUserID string) error {
	return nil
}

// unfollowUser removes a follower from the database
// (Copy from existing code)
func unfollowUser2(userID string, profileUserID string) error {
	return nil
}
