package main

import (
	"database/sql"
	"io/ioutil"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	// # configuration (0) _do we want them public or private?

	var Database string = "./tmp/minitwit.db"
	var Per_page int = 30
	var Debug bool = true
	var Key string = "development key"

	//app aplication ?

	//using db connection (1)
	db, err := connect_db(Database)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//db initialization (2)
	if err := init_db(db, "schema.sql"); err != nil {
		log.Fatal(err)
	}
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
	schema, err := ioutil.ReadFile(schemaFile)
	if err != nil {
		return err
	}

	// Executing the schema SQL after it is being read in the previous step
	_, err = db.Exec(string(schema))
	return err
}

// db query that returns list of dictionaries (3)
func query_db(db *sql.DB, query string, args []interface{}, one bool) ([]map[string]interface{}, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// to have columns and hanling the error
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// to have the result in a slice
	var result []map[string]interface{}
	for rows.Next() {
		//create slice of interface{} to hold the values of the columns
		//another slice to hold a pointer to each item in the interface{} slice
		values := make([]interface{}, len(columns))
		column_pointers := make([]interface{}, len(columns))
		for i, _ := range values {
			column_pointers[i] = &columns[i]
		}

		//results scan into the column pointers
		if err := rows.Scan(column_pointers...); err != nil {
			return nil, err
		}

		//creating  a map to hold the values of the columns, name of the columns as keys
		rowMap := make(map[string]interface{})
		for i, column_name := range columns {
			val := column_pointers[i].(*interface{})
			rowMap[column_name] = *val
		}

		//appending the map to the result slice
		result = append(result, rowMap)

		//if we only want one result, we return the first result
		if one {
			break
		}

	}
	return result, nil
}
