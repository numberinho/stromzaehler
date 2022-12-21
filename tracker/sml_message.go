package tracker

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type SMLMessage struct {
	Message []byte
}

func detectSMLMessage(cache []byte) (SMLMessage, error) {

	startSequence := []byte{0x01b, 0x01b, 0x01b, 0x01b, 0x01, 0x01, 0x01, 0x01}

	startIndex := bytes.Index(cache, startSequence)
	if startIndex == -1 {
		return SMLMessage{}, errors.New("start not found")
	}

	endIndex := bytes.Index(cache[startIndex+8:], startSequence)
	if endIndex == -1 {
		return SMLMessage{}, errors.New("end not found")
	}

	return SMLMessage{cache[startIndex : endIndex+startIndex+8+4]}, nil
}

func (m *SMLMessage) parseKennzahl(k Kennzahl) (float64, error) {

	idx := bytes.Index(m.Message, k.OBIS)
	if idx == -1 {
		return 0, errors.New("error parsing Kennzahl")
	}

	valueByte := m.Message[idx+k.Offset : idx+k.Offset+k.Length]

	for len(valueByte) < 8 { //uint64 = 8bytes
		valueByte = append([]byte{0}, valueByte...)
	}

	return float64(binary.BigEndian.Uint64(valueByte)) / 10000, nil
}

func (m *SMLMessage) parseKennzahlLive(k Kennzahl) (float64, error) {

	idx := bytes.Index(m.Message, k.OBIS)
	if idx == -1 {
		return 0, errors.New("error parsing Kennzahl")
	}

	valueByte := m.Message[idx+k.Offset:]

	idx = bytes.Index(valueByte, []byte{0x01})

	liveSequence := valueByte[:idx]

	/*
		for _, m := range m.Message[idx:] {
			u := fmt.Sprintf("%02x", m)
			fmt.Print(u + " ") //
		}
	*/

	for len(liveSequence) < 8 { //uint64 = 8bytes
		liveSequence = append([]byte{0}, liveSequence...)
	}

	return float64(binary.BigEndian.Uint64(liveSequence)) / 100, nil
}
