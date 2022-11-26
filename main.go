// websockets.go
package main

import (
	"database/sql"
	"zaehler/database"
	"zaehler/tracker"
	"zaehler/ws"
)

var db *sql.DB

func main() {

	wsChannel := make(chan tracker.Zaehlerstand, 1)
	db := database.InitDB()

	go tracker.Tracker.ReadSerialDev(wsChannel, db)

	ws.RunWebserver(wsChannel, db)
}
