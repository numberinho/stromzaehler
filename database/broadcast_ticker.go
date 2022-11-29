package database

import (
	"encoding/json"
	"time"
)

func (db *Database) BroadcastTicker() {
	for range time.Tick(time.Second * 5) {
		x, _ := db.FetchLastHoursNetto(24)
		var output = struct {
			Type   int `json:"Type"`
			Hourly []float64
		}{
			2, x,
		}
		jData, err := json.Marshal(output)
		if err == nil {
			db.BroadcastChannel <- jData
		}
		x, _ = db.FetchLastDaysNetto(8)
		var output2 = struct {
			Type  int `json:"Type"`
			Daily []float64
		}{
			3, x,
		}
		jData, err = json.Marshal(output2)
		if err == nil {
			db.BroadcastChannel <- jData
		}
	}
}
