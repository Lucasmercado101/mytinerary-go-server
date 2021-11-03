package database

import "database/sql"

var Db *sql.DB

type Itinerary struct {
	Id         int
	Title      string
	Creator    int
	Time       string
	Price      string
	Activities []string
	Hashtags   []string
	CityId     int
}
