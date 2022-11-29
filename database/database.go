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
	WriteChannel     chan int
	ReadMutex        sync.Mutex
	BroadcastChannel chan []byte
}

type Datapoint struct {
	Bezug     float64   `json:"Bezug"`
	Abgabe    float64   `json:"Abgabe"`
	Timestamp time.Time `json:"Timestamp"`
}

type YearlyData struct {
	ID    int
	Netto float64
	Data  []MonthlyData
}

type MonthlyData struct {
	ID    int
	Netto float64
	Data  []DailyData
}

type DailyData struct {
	ID    int
	Netto float64
	Data  []HourlyData
}

type HourlyData struct {
	ID    int
	Netto float64
	Data  []Datapoint
}

func InitDB() *Database {
	var db Database
	db.WriteChannel = make(chan int, 1)
	db.WriteChannel <- 1

	db.BroadcastChannel = make(chan []byte, 2)
	go db.BroadcastTicker()
	return &db
}

func (db *Database) fetchHourly(y, m, d, h int) (*HourlyData, error) {
	var hourly HourlyData

	file, err := os.OpenFile(fmt.Sprintf("storage/%d/%d/%d/%d", y, m, d, h), os.O_RDWR, 0660)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&hourly.Data)
	if err != nil {
		return nil, err
	}

	if !(len(hourly.Data) > 0) {
		return nil, errors.New("no data found")
	}

	hourly.Netto = hourly.Data[len(hourly.Data)-1].Bezug - hourly.Data[0].Bezug
	return &hourly, nil
}

func (db *Database) fetchHourlyNetto(y, m, d, h int) (float64, error) {
	hourly, err := db.fetchHourly(y, m, d, h)
	if err != nil {
		return 0, err
	}

	if !(len(hourly.Data) > 0) {
		return 0, errors.New("no data found ")
	}

	return hourly.Data[len(hourly.Data)-1].Bezug - hourly.Data[0].Bezug, nil
}

func (db *Database) FetchLastHoursNetto(n int) ([]float64, error) {
	var hourlySlice = make([]float64, n)

	today := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			targetDate := today.Add(-time.Duration(i) * time.Hour)
			y, m, d := targetDate.Date()

			daily, err := db.fetchHourlyNetto(y, int(m), d, targetDate.Hour())
			if err != nil {
				return
			}
			hourlySlice[i] = daily
		}(i)
	}
	wg.Wait()
	return hourlySlice, nil
}

func (db *Database) fetchDaily(y, m, d int) (*DailyData, error) {
	var daily DailyData

	for h := 0; h < 24; h++ {
		hourly, err := db.fetchHourly(y, m, d, h)
		if err != nil {
			return nil, err
		}
		daily.Data = append(daily.Data, *hourly)
	}
	if !(len(daily.Data)-1 > 0) {
		return nil, errors.New("no data found")
	}
	if !(len(daily.Data[len(daily.Data)-1].Data) > 0) {
		return nil, errors.New("no data found")
	}

	daily.Netto = daily.Data[len(daily.Data)-1].Data[len(daily.Data[len(daily.Data)-1].Data)-1].Bezug - daily.Data[0].Data[0].Bezug
	return &daily, nil
}

func (db *Database) FetchDailyNetto(y, m, d int) (float64, error) {
	var startDaily, endDaily float64

	for h := 0; h < 24; h++ {
		hourly, err := db.fetchHourly(y, m, d, h)
		if err != nil {
			continue
		}
		startDaily = hourly.Data[0].Bezug
		break
	}

	for h := 23; h > -1; h-- {
		hourly, err := db.fetchHourly(y, m, d, h)
		if err != nil {
			continue
		}
		if !(len(hourly.Data) > 0) {
			continue
		}
		endDaily = hourly.Data[len(hourly.Data)-1].Bezug
		break
	}

	return endDaily - startDaily, nil
}

func (db *Database) FetchLastDaysNetto(n int) ([]float64, error) {
	var dailyNettoSlice = make([]float64, n)

	today := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			targetDate := today.AddDate(0, 0, -i)
			y, m, d := targetDate.Date()

			daily, err := db.FetchDailyNetto(y, int(m), d)
			if err != nil {
				return
			}
			dailyNettoSlice[i] = daily
		}(i)
	}
	wg.Wait()
	return dailyNettoSlice, nil
}

func (db *Database) FetchLastDays(n int) ([]DailyData, error) {
	var dailySlice = make([]DailyData, n-1)

	today := time.Now()
	var wg sync.WaitGroup
	for i := 1; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			targetDate := today.AddDate(0, 0, -i)
			y, m, d := targetDate.Date()

			daily, err := db.fetchDaily(y, int(m), d)
			if err != nil {
				return
			}
			dailySlice[i-1] = *daily
		}(i)
	}
	wg.Wait()
	return dailySlice, nil
}

func (db *Database) FetchMonthlyNetto(y, m int) (float64, error) {
	var startMonthly, endMonthly float64

out:
	for d := 0; d <= 31; d++ {
		for h := 1; h < 24; h++ {
			hourly, err := db.fetchHourly(y, m, d, h)
			if err != nil {
				continue
			}
			startMonthly = hourly.Data[0].Bezug
			break out
		}
	}
out2:
	for d := 31; d > 0; d-- {
		for h := 23; h > -1; h-- {
			hourly, err := db.fetchHourly(y, m, d, h)
			if err != nil {
				continue
			}
			endMonthly = hourly.Data[len(hourly.Data)-1].Bezug
			break out2
		}
	}

	return endMonthly - startMonthly, nil
}
