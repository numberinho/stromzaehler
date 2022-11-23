package tracker

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/tarm/serial"
)

type Zaehlerdetail struct {
	Bezug     float64
	Abgabe    float64
	Timestamp time.Time
}

type Zaehlerstand struct {
	Current Zaehlerdetail
	Last    Zaehlerdetail
}

type Kennzahl struct {
	OBIS   []byte
	Offset int
	Length int
}

type SMLMessage struct {
	Message []byte
}

func (z *Zaehlerstand) GetLive() float64 {
	return ((z.Current.Bezug - z.Last.Bezug - z.Current.Abgabe + z.Last.Abgabe) / z.Current.Timestamp.Sub(z.Last.Timestamp).Seconds()) * 3600
}

func (z *Zaehlerstand) writeMySQL(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO zaehlerstand (bezug, abgabe) VALUES (?, ?)", z.Current.Bezug, z.Current.Abgabe)
	if err != nil {
		return fmt.Errorf("addAlbum: %v", err)
	}
	return nil
}

func (z *Zaehlerstand) updateZaehlerstand(bezug, abgabe float64) {
	z.Last = z.Current
	z.Current.Timestamp = time.Now()
	z.Current.Bezug = bezug
	z.Current.Abgabe = abgabe
}

func FetchZaehlerstandUsageToday(db *sql.DB, datum int) (Zaehlerstand, error) {
	var z Zaehlerstand

	row := db.QueryRow("SELECT * FROM zaehlerstand where datum = ?", datum)
	if err := row.Scan(&z.Current.Bezug, &z.Current.Abgabe, &z.Current.Timestamp); err != nil {
		if err == sql.ErrNoRows {
			return z, fmt.Errorf("error %d", datum)
		}
		return z, fmt.Errorf("error %d", datum)
	}
	return z, nil

}

func detectSMLMessage(cache []byte) (SMLMessage, error) {

	startSequence := []byte{0x01b, 0x01b, 0x01b, 0x01b, 0x01, 0x01, 0x01, 0x01}

	startIndex := bytes.Index(cache, startSequence)
	if startIndex == -1 {
		return SMLMessage{}, nil
	}

	endIndex := bytes.Index(cache[startIndex+8:], startSequence)
	if endIndex == -1 {
		return SMLMessage{}, nil
	}

	return SMLMessage{cache[startIndex : endIndex+startIndex+8+4]}, nil
}

func (m *SMLMessage) parseKennzahl(k Kennzahl) (float64, error) {
	idx := bytes.Index(m.Message, k.OBIS)
	if idx == -1 {
		return 0, errors.New("Error parsing Kennzahl")
	}

	valueByte := m.Message[idx+k.Offset : idx+k.Offset+k.Length]

	for len(valueByte) < 8 { //uint64 = 8bytes
		valueByte = append([]byte{0}, valueByte...)
	}

	return float64(binary.BigEndian.Uint64(valueByte)) / 10000, nil
}

func Tracker(wsChannel, dbChannel chan Zaehlerstand, db *sql.DB) {

	var tracker Zaehlerstand

	// Define Kennzahlen
	bezug := Kennzahl{
		OBIS:   []byte{0x77, 0x07, 0x01, 0x00, 0x02, 0x08, 0x00, 0xff},
		Offset: 19,
		Length: 3,
	}

	abgabe := Kennzahl{
		OBIS:   []byte{0x77, 0x07, 0x01, 0x00, 0x01, 0x08, 0x00, 0xff},
		Offset: 22,
		Length: 3,
	}

	// Open serial port
	c := &serial.Config{Name: "/dev/cu.usbserial-0024", Baud: 9600}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize Buffer and Cache
	buf := make([]byte, 128)
	cache := []byte{}

	for {
		// Read buffer
		n, err := s.Read(buf)
		if err != nil {
			log.Fatal(err)
		}

		// Append to cache
		cache = append(cache, buf[:n]...)

		// Search for message
		message, err := detectSMLMessage(cache)

		// If message found, print and clear cache
		if err != nil {

			// Clear cache, we have message now
			cache = []byte{}

			// Parse Kennzahlen
			abgabe, err := message.parseKennzahl(abgabe)
			if err != nil {
				log.Fatal(err)
			}
			bezug, err := message.parseKennzahl(bezug)
			if err != nil {
				log.Fatal(err)
			}
			// Update Zaehlerstand with new Kennzahlen
			tracker.updateZaehlerstand(bezug, abgabe)

			// Write to database
			go tracker.writeMySQL(db)

			// Notificate Sockets
			select {
			case wsChannel <- tracker:
			default:
			}

			select {
			case dbChannel <- tracker:
			default:
			}
		}
	}
}

func NumGen(wsChannel, dbChannel chan Zaehlerstand) {
	var tracker Zaehlerstand

	for {
		time.Sleep(2 * time.Second)

		tracker.Last = tracker.Current
		tracker.Current.Timestamp = time.Now()
		tracker.Current.Abgabe = tracker.Current.Abgabe + rand.Float64()/2
		tracker.Current.Bezug = tracker.Current.Bezug + rand.Float64()

		// send to ws
		select {
		case wsChannel <- tracker:
			fmt.Println("send:", tracker.Current.Abgabe)
		default:
			fmt.Println("!send:", tracker.Current.Abgabe)
		}

		// send to db
		select {
		case dbChannel <- tracker:
		default:
		}
	}
}
