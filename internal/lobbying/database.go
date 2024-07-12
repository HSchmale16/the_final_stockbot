package lobbying

import (
	"database/sql"

	_ "github.com/marcboeker/go-duckdb"
)

/**
 * Contains code to hook into duckdb and run queries.
 */

type LobbyingDB struct {
	DB *sql.DB
}

var LobbyingDBInstance LobbyingDB = LobbyingDB{}

func init() {
	db, err := sql.Open("duckdb", "lobbying.duckdb")
	if err != nil {
		panic(err)
	}

	LobbyingDBInstance.DB = db
}
