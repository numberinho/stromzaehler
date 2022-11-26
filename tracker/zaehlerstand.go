package tracker

import (
	"fmt"
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

func (z *Zaehlerstand) GetLive() float64 {
	return ((z.Current.Bezug - z.Last.Bezug - z.Current.Abgabe + z.Last.Abgabe) / z.Current.Timestamp.Sub(z.Last.Timestamp).Seconds()) * 3600
}

func (z *Zaehlerstand) updateZaehlerstand(bezug, abgabe float64) {
	z.Last = z.Current
	z.Current.Timestamp = time.Now()
	z.Current.Bezug = bezug
	z.Current.Abgabe = abgabe
}

func (z *Zaehlerstand) NotificateWebsocket(wsChannel chan Zaehlerstand) {
	select {
	case wsChannel <- *z:
		fmt.Println("send:", z.Current.Abgabe)
	default:
		fmt.Println("!send:", z.Current.Abgabe)
	}
}
