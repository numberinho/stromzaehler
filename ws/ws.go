package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"zaehler/database"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID   string
	Conn *websocket.Conn
	Pool *Pool
}

type Message struct {
	Type   string `json:"Type"`
	Live   string `json:"Live"`
	Bezug  string `json:"Bezug"`
	Abgabe string `json:"Abgabe"`
	Value  string `json:"Value"`
}

type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (client *Client) HealthCheck() {
	defer func() {
		client.Pool.Unregister <- client
		client.Conn.Close()
	}()

	for {
		_, _, err := client.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func (pool *Pool) Start(db *database.Database) {
	for {
		select {
		//connect
		case client := <-pool.Register:
			pool.Clients[client] = true
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			fmt.Println("Clearing Channel: ", len(pool.Clients))

		//disconnect
		case client := <-pool.Unregister:
			delete(pool.Clients, client)
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))

		//broadcast
		case byteArray := <-db.BroadcastChannel:
			for client := range pool.Clients {
				if err := client.Conn.WriteMessage(1, byteArray); err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
}

func RunWebserver(db *database.Database) {
	pool := NewPool()
	go pool.Start(db)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Fprintf(w, "%+v\n", err)
		}

		client := &Client{
			Conn: conn,
			Pool: pool,
		}

		pool.Register <- client
		client.HealthCheck()

	})

	http.HandleFunc("/history/hourly", hourlyHistory(db))
	http.HandleFunc("/history/daily", dailyHistory(db))

	http.ListenAndServe(":8080", nil)
}

func dailyHistory(db *database.Database) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		x, _ := db.FetchLastHoursNetto(4)
		log.Println(time.Since(start).String())

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		jData, _ := json.Marshal(x)
		w.Write(jData)
	}
}

func hourlyHistory(db *database.Database) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		x, _ := db.FetchLastHoursNetto(24)
		log.Println(time.Since(start).String())

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		jData, _ := json.Marshal(x)
		w.Write(jData)
	}
}
