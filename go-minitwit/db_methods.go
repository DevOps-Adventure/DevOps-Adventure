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
func connect_db(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// to keep separate functions of connection and initialization of the db. (here the db is structured with specific schema/format)
func init_db(db *sql.DB, schemaFile string) error {
	schema, err := os.ReadFile(schemaFile)
	if err != nil {
		return err
	}
	_, err = db.Exec(string(schema))
	return err
}

// query_db executes a query on the database and returns a list of dictionaries (maps)
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
	fmt.Println(result)
	return result, nil
}

/*
	GET DATA
*/

// fetches all public messages for display.
func getPublicMessages(numMsgs int) ([]map[string]interface{}, error) {

	fmt.Println("getPublicMessages")

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
	if numMsgs > 0 {
		args = []interface{}{numMsgs}
	}

	messages, err := query_db(db, query, args, false)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

// fetches all messages from picked user
func getUserMessages(pUserId int64, numMsgs int) ([]map[string]interface{}, error) {

	query := `
	select message.*, user.* from message, user where
    user.user_id = message.author_id and user.user_id = ?
    order by message.pub_date desc limit ?
	`
	args := []interface{}{pUserId, PERPAGE}
	if numMsgs > 0 {
		args = []interface{}{pUserId, numMsgs}
	}
	db, err := connect_db(DATABASE)
	if err != nil {
		return nil, err
	}

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

	db, err := connect_db(DATABASE)
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
func getUserIDByUsername(userName string) (int64, error) {
	fmt.Println("getUserIDByUsername")
	fmt.Println(userName)
	var db, err = connect_db(DATABASE)
	if err != nil {
		fmt.Println("db connection failed")
		return -1, err
	}

	query := `select * from user where username = ?`
	args := []interface{}{userName}
	profile_user, err := query_db(db, query, args, false)

	if profile_user == nil {
		fmt.Println("no profile user")
		return -1, err
	}

	return profile_user[0]["user_id"].(int64), err
}

// fetches a username by their ID
func getUserNameByUserID(userID string) (string, error) {
	var db, err = connect_db(DATABASE)
	if err != nil {
		return "", err
	}

	query := `select * from user where user_id = ?`
	args := []interface{}{userID}
	profile_user, err := query_db(db, query, args, false)

	if profile_user == nil {
		return "", err
	}
	return profile_user[0]["username"].(string), err
}

func getUserByUsername(userName string) ([]map[string]interface{}, error) {
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
func registerUser(userName string, email string, password [16]byte) error {
	query := `insert into user (username, email, pw_hash) values (?, ?, ?)`
	var db, err = connect_db(DATABASE)
	if err != nil {
		return err
	}
	args := []interface{}{userName, email, pq.Array(password)}
	messages, err := query_db(db, query, args, false)
	fmt.Println("this is the messages", messages)
	return err
}

// adds a new message to the database
func addMessage(text string, author_id string) error {
	query := `insert into message (author_id, text, pub_date, flagged) values (?, ?, ?, 0)`
	var db, err = connect_db(DATABASE)
	if err != nil {
		fmt.Println("error in addMessage query")
		return err
	}
	currentTime := time.Now().UTC()
	unixTimestamp := currentTime.Unix()

	args := []interface{}{author_id, text, unixTimestamp, 0}
	query_db(db, query, args, false)
	return err
}

// followUser adds a new follower to the database
func followUser(userID string, profileUserID string) error {
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
func unfollowUser(userID string, profileUserID string) error {
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

// getFollowers fetches up to `limit` followers for the user identified by userID
func getFollowers(userID string, limit int) ([]map[string]interface{}, error) {
	query := `SELECT user.* FROM user 
               INNER JOIN follower ON user.user_id = follower.who_id
               WHERE follower.whom_id = ?
               LIMIT ?`

	var db, err = connect_db(DATABASE)
	if err != nil {
		return nil, err
	}
	args := []interface{}{userID, limit}
	followers, err := query_db(db, query, args, false)
	return followers, err
}

// userExists checks if a user with the given username already exists in the database.
func userExists(username string) (bool, error) {
	query := "SELECT COUNT(*) FROM user WHERE username = ?"
	db, err := connect_db(DATABASE)
	if err != nil {
		return false, err
	}
	defer db.Close()

	var count int
	err = db.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, err
	}

	// 	// If count is greater than 0, the user exists
	// 	return count > 0, nil
	// }
	return true, nil
}
