// websockets.go
package main

import (
	"database/sql"
	"zaehler/tracker"
	"zaehler/ws"
)

var db *sql.DB

func main() {

	wsChannel := make(chan tracker.Zaehlerstand, 1)

	go tracker.Tracker.ReadSerialDev(wsChannel)
	//go tracker.Tracker(wsChannel, dbChannel, db)

	ws.RunWebserver(wsChannel)
}
