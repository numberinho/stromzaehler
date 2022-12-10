package database

import (
	"encoding/json"
	"time"
)

func (db *Database) BroadcastTicker() {
	for range time.Tick(time.Second * 5) {

		// Type 3
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

		// Type 3
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

		// Type 4
		x1, x2, x3, x4, x5, x6, _ := db.FetchCompareWeek()
		var output3 = struct {
			Type         int     `json:"Type"`
			AbgabeThis   float64 `json:"AbgabeThis"`
			BezugThis    float64 `json:"BezugThis"`
			AbgabeLast   float64 `json:"AbgabeLast"`
			BezugLast    float64 `json:"BezugLast"`
			ChangeAbgabe float64 `json:"ChangeAbgabe"`
			ChangeBezug  float64 `json:"ChangeBezug"`
		}{
			4, x1, x2, x3, x4, x5, x6,
		}
		jData, err = json.Marshal(output3)
		if err == nil {
			db.BroadcastChannel <- jData
		}
	}
}
