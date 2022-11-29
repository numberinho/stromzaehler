package tracker

import (
	"time"
)

type Zaehlerdetail struct {
	Bezug     float64   `json:"Bezug"`
	Abgabe    float64   `json:"Abgabe"`
	Timestamp time.Time `json:"Timestamp"`
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

func (z *Zaehlerstand) updateZaehlerstand(bezug, abgabe float64) {
	z.Last = z.Current
	z.Current.Timestamp = time.Now()
	z.Current.Bezug = bezug
	z.Current.Abgabe = abgabe
}
