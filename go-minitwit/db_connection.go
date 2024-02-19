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

/*
	CONNECT, INIT AND QUERY DB
*/

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

/*
	GET DATA
*/

// fetches all public messages for display.
func getPublicMessages() ([]map[string]interface{}, error) {

	query := `
	SELECT message.*, user.* FROM message, user
	WHERE message.flagged = 0 AND message.author_id = user.user_id
	ORDER BY message.pub_date DESC LIMIT ?
	`

	var db, err = connect_db(DATABASE)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	args := []interface{}{PERPAGE}
	messages, err := query_db(db, query, args, false)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

// fetches all messages from picked user
func getUserMessages(pUserId int64) ([]map[string]interface{}, error) {

	query := `
	select message.*, user.* from message, user where
    user.user_id = message.author_id and user.user_id = ?
    order by message.pub_date desc limit ?
	`
	args := []interface{}{pUserId, PERPAGE}
	db, err := connect_db(DATABASE)
	messages, err := query_db(db, query, args, false)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return messages, nil
}

// check whether the given user is followed by logged in
func checkFollowStatus(userID int64, pUserID int64) bool {

	query := `select 1 from follower where
	follower.who_id = ? and follower.whom_id = ?`

	var db, err = connect_db(DATABASE)

	if err != nil {
		return false
	}

	args := []interface{}{userID, pUserID}
	follow_status, err := query_db(db, query, args, false)

	if err != nil {
		return false
	}

	return len(follow_status) > 0
}

// fetches all messages for the current logged in user for 'My Timeline'
func getMyMessages(userID string) ([]map[string]interface{}, error) {
	query := `
    SELECT message.*, user.* FROM message, user
    WHERE message.flagged = 0 AND message.author_id = user.user_id AND (
        user.user_id = ? OR
        user.user_id IN (SELECT whom_id FROM follower WHERE who_id = ?))
    ORDER BY message.pub_date DESC LIMIT ?
    `
	var db, _ = connect_db(DATABASE)
	args := []interface{}{userID, userID, PERPAGE}
	messages, err := query_db(db, query, args, false)

	if err != nil {
		return nil, err
	}

	return messages, nil
}

// fetches a user by their ID
func getUserIDByUsername2(userName string) (int64, error) {
	var db, err = connect_db(DATABASE)
	if err != nil {
		return 0, err
	}

	query := `select * from user where username = ?`
	args := []interface{}{userName}
	profile_user, err := query_db(db, query, args, false)

	if profile_user == nil {
		return 0, err
	}

	return profile_user[0]["user_id"].(int64), err
}

// fetches a username by their ID
func getUserNameByUserID2(userID string) (string, error) {
	var db, err = connect_db(DATABASE)
	if err != nil {
		return "", err
	}

	query := `select * from user where user_id = ?`
	args := []interface{}{userID}
	profile_user, err := query_db(db, query, args, false)

	if profile_user == nil {
		return "no name", err
	}
	return profile_user[0]["username"].(string), err
}

func getUserByUsername2(userName string) ([]map[string]interface{}, error) {
	var db, err = connect_db(DATABASE)
	if err != nil {
		return nil, err
	}

	query := `select * from user where username = ?`
	args := []interface{}{userName}
	profile_user, err := query_db(db, query, args, false)

	if profile_user == nil {
		return nil, err
	}

	return profile_user, err
}

/*
	POST DATA
*/

// registers a new user
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

// adds a new message to the database
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

// followUser adds a new follower to the database
func followUser2(userID string, profileUserID string) error {
	query := `insert into follower (who_id, whom_id) values (?, ?)`
	var db, err = connect_db(DATABASE)
	if err != nil {
		return err
	}
	args := []interface{}{userID, profileUserID}
	messages, err := query_db(db, query, args, false)
	fmt.Println(messages)
	return err
}

// unfollowUser removes a follower from the database
func unfollowUser2(userID string, profileUserID string) error {
	query := `delete from follower where who_id=? and whom_id=?`
	var db, err = connect_db(DATABASE)
	if err != nil {
		return err
	}
	args := []interface{}{userID, profileUserID}
	messages, err := query_db(db, query, args, false)
	fmt.Println(messages)
	return err
}
