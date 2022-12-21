package tracker

import (
	"time"
)

type Zaehlerstand struct {
	Bezug     float64   `json:"Bezug"`
	Abgabe    float64   `json:"Abgabe"`
	Live      float64   `json:"Live"`
	Timestamp time.Time `json:"Timestamp"`
}

type Kennzahl struct {
	OBIS   []byte
	Offset int
	Length int
}

/*
func (z *Zaehlerstand) updateZaehlerstand(bezug, abgabe, live float64) {
	z.Timestamp = time.Now()
	z.Bezug = bezug
	z.Abgabe = abgabe
	z.Live = live
}
*/
