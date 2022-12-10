package tracker

import (
	"log"
	"math/rand"
	"time"
	"zaehler/database"

	"github.com/tarm/serial"
)

type tracker struct {
	Zaehlerstand Zaehlerstand
}

func InitTracker(db *database.Database) tracker {
	var tracker tracker
	go tracker.readSerialDev(db)
	//go tracker.ReadSerial(db)
	return tracker
}

func (t *tracker) ReadSerial(db *database.Database) {

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
			t.Zaehlerstand.updateZaehlerstand(bezug, abgabe)

			// Write to Disc
			go db.Store(t.Zaehlerstand.Current.Abgabe, t.Zaehlerstand.Last.Abgabe, t.Zaehlerstand.Current.Bezug, t.Zaehlerstand.Last.Bezug, t.Zaehlerstand.Current.Timestamp, t.Zaehlerstand.Last.Timestamp)
		}
	}
}

func (t *tracker) readSerialDev(db *database.Database) {

	for {
		time.Sleep(2 * time.Second)

		t.Zaehlerstand.Last = t.Zaehlerstand.Current
		t.Zaehlerstand.Current.Timestamp = time.Now()
		t.Zaehlerstand.Current.Abgabe = t.Zaehlerstand.Current.Abgabe + rand.Float64()/2
		t.Zaehlerstand.Current.Bezug = t.Zaehlerstand.Current.Bezug + rand.Float64()

		go db.Store(t.Zaehlerstand.Current.Abgabe, t.Zaehlerstand.Last.Abgabe, t.Zaehlerstand.Current.Bezug, t.Zaehlerstand.Last.Bezug, t.Zaehlerstand.Current.Timestamp, t.Zaehlerstand.Last.Timestamp)
	}
}
