package database

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func (db *Database) Store(a, at, b, bt float64, t, tt time.Time) {

	y, m, d := t.Date()
	h := t.Hour()
	directory := fmt.Sprintf("storage/%d/%d/%d", y, m, d)

	_, err := os.Stat(directory)
	if os.IsNotExist(err) {
		os.MkdirAll(directory, os.ModePerm)
	}

	<-db.WriteChannel

	file, err := os.OpenFile(fmt.Sprintf("%s/%d", directory, h), os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		fmt.Println("store", err)
	}
	defer file.Close()

	var read []Datapoint

	decoder := gob.NewDecoder(file)
	decoder.Decode(&read)

	read = append(read, Datapoint{a, b, t})

	file.Seek(0, 0)
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(read)
	if err != nil {
		panic(err)
	}
	db.WriteChannel <- 1

	var output = struct {
		Type   int     `json:"Type"`
		Abgabe float64 `json:"Abgabe"`
		Bezug  float64 `json:"Bezug"`
		Live   float64 `json:Live`
	}{
		1, a, b, (b - bt - a + at) / t.Sub(tt).Seconds() * 3600,
	}

	byteArray, err := json.Marshal(output)
	if err != nil {
		return
	}

	select {
	case db.BroadcastChannel <- byteArray:
		fmt.Print("_")
	default:
		fmt.Print("!")
	}
}
