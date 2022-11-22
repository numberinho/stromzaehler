// websockets.go
package main

import (
	"time"
	"zaehler/tracker"
	"zaehler/ws"
)

func main() {

	wsChannel := make(chan tracker.Zaehlerstand, 1)
	dbChannel := make(chan tracker.Zaehlerstand, 1)

	go tracker.NumGen(wsChannel, dbChannel)
	go ws.RunWebserver(wsChannel)

	time.Sleep(1000 * time.Second)
}
