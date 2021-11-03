package database

import (
	"database/sql"

	"github.com/lib/pq"
)

var Db *sql.DB

type Itinerary struct {
	Id         int
	Title      string
	Creator    int
	Time       string
	Price      string
	Activities pq.StringArray
	Hashtags   pq.StringArray
	CityId     int
}
