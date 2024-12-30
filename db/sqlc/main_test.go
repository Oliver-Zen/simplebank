package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/techschool/simplebank/util"
)

// const (
// 	dbDriver = "postgres" // Specifies the database driver to use
// 	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
// )

var testQueries *Queries // Global variable to hold the Queries instance for tests
var testDB *sql.DB       // Global variable to hold the database connection instance

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../../")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	
	// Establish a database connection
	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to the db:", err)
	}

	// Initialize the Queries instance using the database connection
	testQueries = New(testDB)

	// Execute all tests and exit with the appropriate status code
	os.Exit(m.Run())
}
