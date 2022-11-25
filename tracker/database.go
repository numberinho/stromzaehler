package tracker

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type database struct {
	Mutex chan int
}

var DB database

func init() {
	DB.Mutex = make(chan int, 1)
	DB.Mutex <- 1
}

func (t *tracker) Store() {

	y, m, d := t.Zaehlerstand.Current.Timestamp.Date()
	h := t.Zaehlerstand.Current.Timestamp.Hour()
	directory := fmt.Sprintf("storage/%d/%d/%d", y, m, d)

	_, err := os.Stat(directory)
	if os.IsNotExist(err) {
		os.MkdirAll(directory, os.ModePerm)
	}

	<-DB.Mutex

	file, err := os.OpenFile(fmt.Sprintf("%s/%d", directory, h), os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		fmt.Println("store", err)
	}
	defer file.Close()

	var read []Zaehlerdetail

	decoder := gob.NewDecoder(file)
	decoder.Decode(&read)

	read = append(read, t.Zaehlerstand.Current)

	file.Seek(0, 0)
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(read)
	if err != nil {
		panic(err)
	}

	// message ready
	DB.Mutex <- 1
}

func FetchDay(dirname string) (*[]Zaehlerdetail, error) {
	var day []Zaehlerdetail

	var paths []string
	err := filepath.WalkDir(dirname, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			paths = append(paths, path)
			return nil
		}
		return err
	})
	if err != nil {
		return &day, err
	}

	for _, path := range paths {
		var hour []Zaehlerdetail

		file, err := os.OpenFile(path, os.O_RDWR, 0660)
		defer file.Close()
		if err != nil {
			continue
		}
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(&hour)
		if err != nil {
			continue
		}

		day = append(day, hour...)
	}
	if len(day) > 0 {
		return &day, nil
	}
	return &day, errors.New("no data found")
}

func FetchLastN(n int) (map[string]([]Zaehlerdetail), error) {
	var mutex = &sync.Mutex{}
	var lastNdays = make(map[string]([]Zaehlerdetail))

	today := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			thisDate := today.AddDate(0, 0, -i)
			y, m, d := thisDate.Date()

			directory := fmt.Sprintf("storage/%d/%d/%d", y, m, d)

			day, err := FetchDay(directory)
			if err != nil {
				return
			}
			if len(*day) > 0 {
				mutex.Lock()
				lastNdays[directory] = *day
				mutex.Unlock()
			}
		}(i)
	}
	wg.Wait()
	return lastNdays, nil
}
