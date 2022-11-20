package main

import (
	"bytes"
	"encoding/binary"
	"log"

	"github.com/tarm/serial"
)

func main() {

	startSequence := []byte{0x01b, 0x01b, 0x01b, 0x01b, 0x01, 0x01, 0x01, 0x01}

	c := &serial.Config{Name: "/dev/cu.usbserial-0024", Baud: 9600}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, 128)
	cache := []byte{}

	for {
		// read buffer
		n, err := s.Read(buf)
		if err != nil {
			log.Fatal(err)
		}

		// append to cache
		cache = append(cache, buf[:n]...)

		// search for message
		message := getMessage(cache, startSequence)

		// if message found, print and clear cache
		if message != nil {
			//log.Printf("message: %x \n", message)
			//log.Printf("message: %q \n", message)

			parseMessage(message)
			cache = []byte{}
		}
	}
}

func parseMessage(message []byte) {
	OBIS1 := []byte{0x77, 0x07, 0x01, 0x00, 0x02, 0x08, 0x00, 0xff}
	OBIS2 := []byte{0x77, 0x07, 0x01, 0x00, 0x01, 0x08, 0x00, 0xff}

	idx1 := bytes.Index(message, OBIS1)
	if idx1 == -1 {
		return
	}
	idx1 += 19

	idx2 := bytes.Index(message, OBIS2)
	if idx2 == -1 {
		return
	}
	idx2 += 22

	value1byte := message[idx1 : idx1+3]
	value2byte := message[idx2 : idx2+3]

	for len(value1byte) < 8 { //uint64 = 8bytes
		value1byte = append([]byte{0}, value1byte...)
	}

	for len(value2byte) < 8 { //uint64 = 8bytes
		value2byte = append([]byte{0}, value2byte...)
	}

	value1 := float32(binary.BigEndian.Uint64(value1byte)) / 10000
	value2 := float32(binary.BigEndian.Uint64(value2byte)) / 10000

	log.Printf("ABGABE: %f kWh\n", value1)
	log.Printf("BEZUG: %f kWh\n", value2)

}

func getMessage(cache, startSequence []byte) []byte {
	startIndex := bytes.Index(cache, startSequence)
	if startIndex == -1 {
		//log.Println("No start found")
		return nil
	}
	//fmt.Println("message start:", startIndex+4)

	endIndex := bytes.Index(cache[startIndex+8:], startSequence)
	if endIndex == -1 {
		//log.Println("No end found")
		return nil
	}
	//fmt.Println("message end:", endIndex+startIndex+4)

	return cache[startIndex : endIndex+startIndex+8+4]
}
