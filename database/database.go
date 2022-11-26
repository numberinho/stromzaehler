package database

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

type Database struct {
	WriteChannel chan int
	ReadMutex    sync.Mutex
}

type Datapoint struct {
	Bezug     float64   `json:"Bezug"`
	Abgabe    float64   `json:"Abgabe"`
	Timestamp time.Time `json:"Timestamp"`
}

func InitDB() *Database {
	var db Database
	db.WriteChannel = make(chan int, 1)
	db.WriteChannel <- 1
	return &db
}

func (db *Database) Store(a, b float64, t time.Time) {

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

	// message ready
	db.WriteChannel <- 1
}

func (db *Database) fetchDay(dirname string) (*[]Datapoint, error) {
	var day []Datapoint

	// TODO: goroutine?
	for i := 0; i < 24; i++ {
		var hour []Datapoint

		file, err := os.OpenFile(fmt.Sprintf("%s/%d", dirname, i), os.O_RDWR, 0660)
		defer file.Close()
		if err != nil {
			break
		}
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(&hour)
		if err != nil {
			break
		}
		day = append(day, hour...)
	}
	if len(day) > 0 {
		return &day, nil
	}
	return &day, errors.New("no data found")
}

func (db *Database) fetchLastN(n int) (map[int]([]Datapoint), error) {
	var lastNdays = make(map[int]([]Datapoint))

	today := time.Now()
	var wg sync.WaitGroup
	for i := 1; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			thisDate := today.AddDate(0, 0, -i)
			y, m, d := thisDate.Date()

			directory := fmt.Sprintf("storage/%d/%d/%d", y, m, d)

			day, err := db.fetchDay(directory)
			if err != nil {
				return
			}
			if len(*day) > 0 {
				db.ReadMutex.Lock()
				lastNdays[i] = *day
				db.ReadMutex.Unlock()
			}
		}(i)
	}
	wg.Wait()
	return lastNdays, nil
}

func (db *Database) FetchLastNDailyData(n int) ([]Datapoint, error) {
	lastNdays, err := db.fetchLastN(n)
	if err != nil {
		return nil, err
	}
	arr := make([]Datapoint, n)
	for k, v := range lastNdays {
		arr[k].Bezug = v[len(v)-1].Bezug - v[0].Bezug
		arr[k].Abgabe = v[len(v)-1].Bezug - v[0].Bezug
		arr[k].Timestamp = v[0].Timestamp
	}
	return arr, nil
}
