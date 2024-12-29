package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres" // Specifies the database driver to use
	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" 
)

var testQueries *Queries // Global variable to hold the Queries instance for tests
var testDB *sql.DB       // Global variable to hold the database connection instance

func TestMain(m *testing.M) {
	
	// Establish a database connection
	var err error
	testDB, err = sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to the db:", err)
	}

	// Initialize the Queries instance using the database connection
	testQueries = New(testDB)

	// Execute all tests and exit with the appropriate status code
	os.Exit(m.Run())
}
