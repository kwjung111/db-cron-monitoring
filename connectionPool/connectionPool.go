package connectionPool

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var connection *sql.DB

var dsn string

// initialize and return database connection pool
func initPool() *sql.DB {
	if connection != nil {
		return connection
	}

	pool, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Error opening db connection ", err)
	}

	// verify connection
	err = pool.Ping()
	if err != nil {
		log.Fatal("Error verify connection", err)
	}

	pool.SetConnMaxLifetime(time.Minute * 3)
	pool.SetMaxOpenConns(5)
	pool.SetMaxIdleConns(5)

	connection = pool

	return pool
}

// returns database connection pool (singleton)
func GetConnection() *sql.DB {
	if connection == nil {
		return initPool()
	}

	return connection
}

func SetDsn(DSN string) {
	dsn = DSN
}
