// websockets.go
package main

import (
	"zaehler/database"
	"zaehler/tracker"
	"zaehler/ws"
)

func main() {

	wsChannel := make(chan tracker.Zaehlerstand, 1)
	db := database.InitDB()

	go tracker.Tracker.ReadSerialDev(wsChannel, db)

	ws.RunWebserver(wsChannel, db)
}
