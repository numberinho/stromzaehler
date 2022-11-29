// websockets.go
package main

import (
	"zaehler/database"
	"zaehler/tracker"
	"zaehler/ws"
)

func main() {

	db := database.InitDB()
	tracker := tracker.InitTracker(db)

	go tracker.ReadSerialDev(db)
	ws.RunWebserver(db)
}
