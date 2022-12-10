// websockets.go
package main

import (
	"zaehler/database"
	"zaehler/tracker"
	"zaehler/ws"
)

func main() {

	db := database.InitDB()
	tracker.InitTracker(db)

	ws.RunWebserver(db)
}
