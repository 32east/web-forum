package sqlDb

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"os"
)

var MySqlDB *sql.DB

func ConnectDatabase() *sql.DB {
	config := mysql.Config{
		User:                 os.Getenv("MYSQL_USER"),
		Passwd:               os.Getenv("MYSQL_PASS"),
		Net:                  os.Getenv("MYSQL_NET"),
		Addr:                 os.Getenv("MYSQL_ADDR"),
		DBName:               os.Getenv("MYSQL_DATABASE"),
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	Database, err := sql.Open("mysql", config.FormatDSN())

	if err != nil {
		panic(err)
	}

	err = Database.Ping()

	if err != nil {
		panic(err)
	}

	MySqlDB = Database
	return Database
}
