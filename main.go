// websockets.go
package main

import (
	"database/sql"
	"time"
	"zaehler/tracker"
	"zaehler/ws"
)

var db *sql.DB

func main() {

	wsChannel := make(chan tracker.Zaehlerstand, 1)
	dbChannel := make(chan tracker.Zaehlerstand, 1)

	//db = database.ConnectDB()

	go tracker.NumGen(wsChannel, dbChannel)
	//go tracker.Tracker(wsChannel, dbChannel, db)

	go ws.RunWebserver(wsChannel)

	time.Sleep(5000 * time.Second)
}
