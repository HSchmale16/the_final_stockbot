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
	// why do I have to open read only I don't understand?
	db, err := sql.Open("duckdb", "lobbying.duckdb?access_mode=read_only&threads=4")
	if err != nil {
		panic(err)
	}

	LobbyingDBInstance.DB = db
}
